import asyncio
import argparse
import json
import os
import sys
import uvicorn
import logging
from typing import Any, List, Optional, Dict, Callable
import requests
import uuid
import webbrowser
from threading import Thread
from datetime import datetime
import base64

# OpenAI client for DashScope
from openai import OpenAI

# Required imports for browser-use
from browser_use import Agent, Controller, ActionResult
from browser_use.browser.browser import Browser, BrowserConfig
from dotenv import load_dotenv

# Required imports for custom model adapter
from langchain_core.language_models.chat_models import BaseChatModel
from langchain_core.messages import AIMessage, BaseMessage, HumanMessage, SystemMessage
from langchain_core.outputs import ChatGeneration, ChatResult

# API server imports
from fastapi import FastAPI, WebSocket, WebSocketDisconnect, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.staticfiles import StaticFiles

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("integrated-agent")

load_dotenv()

# Custom model adapter for Qwen Vision using DashScope
class QwenVisionModel(BaseChatModel):
    """Custom LLM adapter for Qwen VL model via DashScope"""
    
    def __init__(self, api_key: Optional[str] = None, model_name: str = "qwen-vl-plus", max_tokens: int = 1024, **kwargs):
        super().__init__(**kwargs)
        self.api_key = api_key or os.environ.get("DASHSCOPE_API_KEY", "sk-d199dc789f9d45d7b29a459eb2169c35")
        self.model_name = model_name
        self.max_tokens = max_tokens
        self.client = OpenAI(
            api_key=self.api_key,
            base_url="https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
        )
        logger.info(f"QwenVisionModel initialized with model: {model_name}")
    
    def _generate(self, messages: List[BaseMessage], stop: Optional[List[str]] = None, 
                 run_manager: Optional[Any] = None, **kwargs) -> ChatResult:
        """Connect to the Qwen model using DashScope API"""
        try:
            # Convert LangChain messages to OpenAI format
            openai_messages = []
            
            for message in messages:
                if isinstance(message, HumanMessage):
                    role = "user"
                elif isinstance(message, AIMessage):
                    role = "assistant"
                elif isinstance(message, SystemMessage):
                    role = "system"
                else:
                    role = "user"
                
                content = message.content
                
                # Check if content has images (e.g., from browser screenshot)
                if isinstance(content, str) and "data:image/" in content:
                    # Extract all images from the content
                    text_parts = []
                    image_parts = []
                    
                    # Simple parsing for images in markdown format: ![alt](data:image/...)
                    import re
                    image_pattern = r'!\[(.*?)\]\((data:image/[^;]+;base64,[a-zA-Z0-9+/=]+)\)'
                    matches = re.findall(image_pattern, content)
                    
                    # Extract text content without images
                    clean_content = re.sub(image_pattern, '', content).strip()
                    if clean_content:
                        text_parts.append(clean_content)
                    
                    # Add image parts
                    for _, img_data in matches:
                        image_parts.append(img_data)
                    
                    # Create multimodal message content
                    msg_content = []
                    if text_parts:
                        msg_content.append({"type": "text", "text": "\n".join(text_parts)})
                    
                    for img_data in image_parts:
                        msg_content.append({
                            "type": "image_url",
                            "image_url": {"url": img_data}
                        })
                    
                    openai_messages.append({"role": role, "content": msg_content})
                else:
                    # Regular text message
                    openai_messages.append({"role": role, "content": content})
            
            # Call the DashScope API
            response = self.client.chat.completions.create(
                model=self.model_name,
                messages=openai_messages,
                max_tokens=self.max_tokens
            )
            
            # Extract the response content
            message_content = response.choices[0].message.content
            
            # Create a LangChain-compatible response
            message = AIMessage(content=message_content)
            generation = ChatGeneration(message=message)
            
            return ChatResult(generations=[generation])
            
        except Exception as e:
            logger.error(f"Error in model generation: {str(e)}")
            # Return a fallback response
            message = AIMessage(content=f"Error: {str(e)}")
            generation = ChatGeneration(message=message)
            return ChatResult(generations=[generation])
    
    @property
    def _llm_type(self) -> str:
        return "qwen-vision-dashscope"


# Configure CORS for web UI
app = FastAPI()

# Configure CORS for web UI
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Allow all origins for flexibility
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Store active sessions
active_sessions = {}

class AgentSession:
    def __init__(self, api_key=None, model_name="qwen-vl-plus", headless=True):
        self.model = QwenVisionModel(api_key=api_key, model_name=model_name)
        self.browser = Browser(config=BrowserConfig(headless=headless, disable_security=True))
        self.controller = Controller()
        self.agent = None
        self.is_running = False
        self.status = "initialized"

    async def cleanup(self):
        if self.browser:
            await self.browser.close()
            self.browser = None

@app.websocket("/ws/{session_id}")
async def websocket_endpoint(websocket: WebSocket, session_id: str):
    await websocket.accept()
    
    # Get API key and model from query parameters or use defaults
    api_key = websocket.query_params.get("api_key")
    model_name = websocket.query_params.get("model_name", "qwen-vl-plus")
    headless = websocket.query_params.get("headless", "true").lower() == "true"
    
    try:
        # Initialize session
        session = AgentSession(api_key=api_key, model_name=model_name, headless=headless)
        active_sessions[session_id] = session
        
        await websocket.send_json({"type": "status", "message": "Ready", "session": session_id})
        logger.info(f"Session {session_id} initialized")
        
        # Main communication loop
        while True:
            data = await websocket.receive_text()
            data = json.loads(data)
            
            if data["type"] == "task":
                # Handle task request
                if session.is_running:
                    await websocket.send_json({"type": "error", "message": "Agent already running"})
                    continue
                
                # Get task details
                task = data["content"]
                max_steps = int(data.get("max_steps", 30))
                use_vision = data.get("use_vision", True)
                
                logger.info(f"Session {session_id} starting task: {task[:50]}...")
                
                # Create agent
                session.agent = Agent(
                    task=task,
                    llm=session.model,
                    controller=session.controller,
                    browser=session.browser,
                    use_vision=use_vision
                )
                
                # Run agent in background task
                session.is_running = True
                asyncio.create_task(run_agent_task(websocket, session_id, session, max_steps))
                
            elif data["type"] == "stop":
                # Stop the agent
                if session.agent and session.is_running:
                    session.agent.state.stopped = True
                    session.is_running = False
                    await websocket.send_json({"type": "status", "message": "Agent stopped"})
                    logger.info(f"Session {session_id} stopped")
                else:
                    await websocket.send_json({"type": "status", "message": "No agent running"})
    
    except WebSocketDisconnect:
        logger.info(f"WebSocket disconnected for session {session_id}")
        # Clean up resources
        if session_id in active_sessions:
            await active_sessions[session_id].cleanup()
            del active_sessions[session_id]
    
    except Exception as e:
        logger.error(f"Error in WebSocket handler: {str(e)}")
        await websocket.send_json({"type": "error", "message": str(e)})
        if session_id in active_sessions:
            await active_sessions[session_id].cleanup()
            del active_sessions[session_id]

async def run_agent_task(websocket, session_id, session, max_steps):
    try:
        # Initialize step tracking
        last_step = -1
        
        # Create step observer
        async def step_observer(step_number, history):
            nonlocal last_step
            if step_number > last_step:
                last_step = step_number
                # Send step update
                await websocket.send_json({
                    "type": "step",
                    "step_number": step_number,
                    "action": history.last_action() if history.last_action() else "Thinking...",
                    "result": history.last_result() if history.last_result() else ""
                })
                logger.debug(f"Session {session_id} step {step_number}")
        
        # Add observer to agent
        session.agent.add_observer(step_observer)
        
        # Run the agent
        await websocket.send_json({"type": "status", "message": "Agent running"})
        history = await session.agent.run(max_steps=max_steps)
        
        # Send final result
        final_result = history.final_result() or "Task completed but no final result provided."
        await websocket.send_json({
            "type": "complete",
            "result": final_result
        })
        logger.info(f"Session {session_id} completed")
        
    except Exception as e:
        logger.error(f"Error running agent: {str(e)}")
        await websocket.send_json({"type": "error", "message": str(e)})
    finally:
        session.is_running = False

# Health check endpoint
@app.get("/health")
async def health_check():
    return {"status": "ok", "timestamp": datetime.now().isoformat()}

# API endpoint to create new session
@app.post("/api/session")
async def create_session():
    session_id = str(uuid.uuid4())
    return {"session_id": session_id}

# Command-line interface
async def run_cli(api_key, model_name, headless, max_steps):
    # Get task from user
    task = input("Enter your task: ")
    
    # Create model instance
    model = QwenVisionModel(api_key=api_key, model_name=model_name)
    
    # Create browser, controller, and agent
    browser = Browser(config=BrowserConfig(headless=headless, disable_security=True))
    controller = Controller()
    
    # Add any custom actions if needed
    @controller.action('Take screenshot')
    async def take_screenshot(browser=None):
        if browser:
            page = await browser.get_current_page()
            screenshot = await page.screenshot()
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            file_path = f"screenshot_{timestamp}.png"
            with open(file_path, "wb") as f:
                f.write(screenshot)
            return ActionResult(extracted_content=f"Screenshot saved to {file_path}")
        return ActionResult(extracted_content="Browser not available")
    
    # Create agent
    agent = Agent(
        task=task, 
        llm=model, 
        controller=controller, 
        browser=browser, 
        use_vision=True
    )
    
    # Run agent
    try:
        print(f"\nRunning task: {task}\n")
        history = await agent.run(max_steps=max_steps)
        print("\n" + "="*50)
        print("Task completed. Final result:")
        print("="*50)
        print(history.final_result() or "No final result provided.")
    finally:
        await browser.close()

def start_server(host, port):
    """Start the API server in a separate thread"""
    uvicorn.run(app, host=host, port=port)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Integrated Agent with Browser-use and Qwen Vision")
    parser.add_argument("--mode", choices=["server", "cli"], default="server", help="Run as API server or CLI")
    parser.add_argument("--api-key", help="API key for DashScope")
    parser.add_argument("--model-name", default="qwen-vl-plus", help="Model name (qwen-vl-plus, etc.)")
    parser.add_argument("--host", default="127.0.0.1", help="Host for the API server")
    parser.add_argument("--port", type=int, default=8000, help="Port for the API server")
    parser.add_argument("--headless", action="store_true", help="Run browser in headless mode")
    parser.add_argument("--max-steps", type=int, default=30, help="Maximum steps for agent execution")
    parser.add_argument("--open-ui", action="store_true", help="Open web UI in browser (if available)")
    
    args = parser.parse_args()
    
    if args.mode == "server":
        logger.info(f"Starting server on {args.host}:{args.port}")
        print(f"Starting API server at http://{args.host}:{args.port}")
        print(f"WebSocket endpoint: ws://{args.host}:{args.port}/ws/{{session_id}}")
        
        if args.open_ui:
            # Try to open web UI if running
            web_ui_url = "http://127.0.0.1:7788"
            try:
                # Check if web UI is running
                response = requests.get(web_ui_url, timeout=1)
                if response.status_code == 200:
                    webbrowser.open(web_ui_url)
                    print(f"Opening web UI at {web_ui_url}")
            except:
                print("Web UI not detected at http://127.0.0.1:7788")
        
        # Start server
        start_server(args.host, args.port)
    else:
        # Run CLI
        print("Starting CLI mode...")
        asyncio.run(run_cli(args.api_key, args.model_name, args.headless, args.max_steps))