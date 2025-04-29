from __future__ import annotations
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, List, Optional, Type, TypeVar, Generic, Any
import os
import tkinter as tk
from tkinter import ttk, messagebox
import requests
import http.server
import socketserver
import threading
import json
import logging
import asyncio
from enum import Enum
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from pathlib import Path
from dotenv import load_dotenv
from pydantic import BaseModel, Field, ValidationError
from functools import wraps

from openai import OpenAI
from typing import List, Optional, Dict, Any
from datetime import datetime
from dataclasses import dataclass
from openai import AzureOpenAI
from pydantic import BaseModel, Field
from azure.ai.inference import ChatCompletionsClient
from azure.ai.inference.models import AssistantMessage, SystemMessage, UserMessage
from azure.core.credentials import AzureKeyCredential
import concurrent
from alibabahandler import AlibabaHandler
from deepinfrahandler import DeepInfraHandler
from githubazure import GithubAzureHandler
from nvidiahandler import NvidiaHandler
from base_types import ModelProvider, BaseProviderHandler


# Type variables for generic types
HandlerT = TypeVar('HandlerT', bound='BaseProviderHandler')

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - [%(threadName)s] - %(message)s'
)
logger = logging.getLogger(__name__)

class ConfigurationError(Exception):
    """Raised when configuration loading or validation fails"""
    pass

class ProviderError(Exception):
    """Base exception for provider-related errors"""
    pass

class ModelFetchError(ProviderError):
    """Raised when fetching models fails"""
    pass

class RequestProcessingError(ProviderError):
    """Raised when processing a request fails"""
    pass

@dataclass
class EnvironmentConfig:
    """Environment configuration with validation"""
    alibaba_api_key: Optional[str] = None
    deepinfra_api_key: Optional[str] = None
    github_token: Optional[str] = None
    nvidia_api_key: Optional[str] = None
    
    @classmethod
    def from_env(cls) -> 'EnvironmentConfig':
        """Load configuration from environment variables"""
        load_dotenv(Path('.env'))
        
        return cls(
            alibaba_api_key=os.getenv('DASHSCOPE_API_KEY'),
            deepinfra_api_key=os.getenv('DEEPINFRA_API_KEY'),
            github_token=os.getenv('GITHUB_TOKEN'),
            nvidia_api_key=os.getenv('NVIDIA_API_KEY'),

        )

class ChatMessage(BaseModel):
    """Base model for chat messages with role and content validation."""
    role: str = Field(..., pattern="^(system|user|assistant)$")
    content: str = Field(..., min_length=1)

class ChatConfig(BaseModel):
    """Configuration settings for the chat client."""
    endpoint: str
    api_key: str
    deployment_name: str
    api_version: str = "2024-02-15-preview"

class AzureChatClient:
    """Manages communication with Azure OpenAI's chat completions API."""
    
    def __init__(self, config: ChatConfig) -> None:
        """
        Initialize the Azure Chat Client with configuration.
        
        Args:
            config (ChatConfig): Validated configuration settings
        """
        self.config = config
        self.client = AzureOpenAI(
            azure_endpoint=config.endpoint,
            api_key=config.api_key,
            api_version=config.api_version
        )
        logger.info(f"Initialized Azure Chat Client for deployment: {config.deployment_name}")

    async def send_message(
        self,
        messages: List[ChatMessage],
        temperature: float = 0.7,
        max_tokens: Optional[int] = None
    ) -> Dict[str, Any]:
        """
        Send messages to Azure OpenAI and get completion.
        
        Args:
            messages (List[ChatMessage]): List of validated chat messages
            temperature (float, optional): Sampling temperature. Defaults to 0.7
            max_tokens (Optional[int], optional): Maximum tokens in response
            
        Returns:
            Dict[str, Any]: Completion response from Azure OpenAI
        """
        try:
            response = await self.client.chat.completions.create(
                model=self.config.deployment_name,
                messages=[msg.dict() for msg in messages],
                temperature=temperature,
                max_tokens=max_tokens
            )
            logger.info("Successfully received chat completion")
            return response
            
        except Exception as e:
            logger.error(f"Error in chat completion request: {str(e)}")
            raise

# Helper functions for message creation
def system_message(content: str) -> ChatMessage:
    return ChatMessage(role="system", content=content)

def user_message(content: str) -> ChatMessage:
    return ChatMessage(role="user", content=content)

def assistant_message(content: str) -> ChatMessage:
    return ChatMessage(role="assistant", content=content)

class ModelProvider(Enum):
    ALIBABA = "Alibaba"
    DEEPINFRA = "DeepInfra"
    GITHUB_AZURE = "GitHub Pat"
    NVIDIA = "Nvidia"

    
    @property
    def api_key_name(self) -> str:
        """Get the corresponding environment variable name for the provider"""
        mapping = {
            self.ALIBABA: "alibaba_api_key",
            self.DEEPINFRA: "deepinfra_api_key",
            self.GITHUB_AZURE: "github_token",
            self.NVIDIA: "nvidia_api_key",

        }
        return mapping[self]



def log_operation(func):
    """Decorator for logging method calls with timing"""
    @wraps(func)
    async def wrapper(*args, **kwargs):
        start_time = datetime.now()
        try:
            result = await func(*args, **kwargs)
            elapsed = (datetime.now() - start_time).total_seconds()
            logger.info(f"{func.__name__} completed in {elapsed:.2f}s")
            return result
        except Exception as e:
            elapsed = (datetime.now() - start_time).total_seconds()
            logger.error(f"{func.__name__} failed after {elapsed:.2f}s: {str(e)}")
            raise
    return wrapper

@dataclass
class ModelInfo:
    """Model information container"""
    id: str
    provider: ModelProvider
    capabilities: List[str]



class EnhancedAPIRouter:
    """Enhanced API router with environment configuration support"""
    def __init__(self):
        self.env_config = EnvironmentConfig.from_env()
        self.current_handler: Optional[BaseProviderHandler] = None
        self.server: Optional[socketserver.ThreadingTCPServer] = None
        self.server_thread: Optional[threading.Thread] = None
        self.running = False
        self.port = 8000
        self._handlers: Dict[ModelProvider, Type[BaseProviderHandler]] = {
            ModelProvider.ALIBABA: AlibabaHandler,
            ModelProvider.DEEPINFRA: DeepInfraHandler,
            ModelProvider.GITHUB_AZURE: GithubAzureHandler,
            ModelProvider.NVIDIA: NvidiaHandler

        }

    def get_api_key(self, provider: ModelProvider) -> Optional[str]:
        """Get API key for specified provider from constants"""
        key_mapping = {
            ModelProvider.ALIBABA: DASHSCOPE_API_KEY,
            ModelProvider.DEEPINFRA: DEEPINFRA_API_KEY, 
            ModelProvider.GITHUB_AZURE: GITHUB_TOKEN,
            ModelProvider.NVIDIA: NVIDIA_API_KEY
        }
        return key_mapping.get(provider)

    def create_request_handler(self):
        """Create request handler class with router reference"""
        router = self

        class RequestHandler(http.server.SimpleHTTPRequestHandler):
            async def _handle_request(self):
                content_length = int(self.headers.get('Content-Length', 0))
                post_data = self.rfile.read(content_length)
                request_data = json.loads(post_data.decode('utf-8'))
                
                # Override the model with currently selected one
                if 'model' in request_data:
                    selected_model = router.current_handler.model_name
                    request_data['model'] = selected_model
                
                try:
                    response_data = await router.current_handler.process_request(
                        self.path, request_data
                    )
                    
                    self.send_response(200)
                    self.send_header('Content-Type', 'application/json')
                    self.end_headers()
                    self.wfile.write(json.dumps(response_data).encode('utf-8'))
                except RequestProcessingError as e:
                    self.send_error(500, str(e))
                except Exception as e:
                    logger.error(f"Unexpected error: {e}")
                    self.send_error(500, "Internal server error")

            def do_POST(self):
                asyncio.run(self._handle_request())

        return RequestHandler

    def start_server(self, provider: ModelProvider) -> bool:
        """Start server with specified provider"""
        if self.running:
            return False
            
        api_key = self.get_api_key(provider)
        if not api_key:
            raise ConfigurationError(f"No API key found for {provider.value}")
            
        handler_class = self._handlers.get(provider)
        if not handler_class:
            raise ValueError(f"Unsupported provider: {provider}")
            
        self.current_handler = handler_class(api_key)
        
        try:
            self.server = socketserver.ThreadingTCPServer(
                ("", self.port),
                self.create_request_handler()
            )
            
            self.server_thread = threading.Thread(
                target=self.server.serve_forever,
                name=f"{provider.value}ServerThread"
            )
            self.server_thread.daemon = True
            self.server_thread.start()
            
            self.running = True
            logger.info(f"Server started for {provider.value} on port {self.port}")
            return True
        except Exception as e:
            logger.error(f"Failed to start server: {e}")
            self.stop_server()
            raise

    def stop_server(self) -> bool:
        """Stop the running server"""
        if not self.running:
            return False
            
        try:
            self.server.shutdown()
            self.server.server_close()
            self.server_thread.join()
            self.running = False
            logger.info("Server stopped")
            return True
        except Exception as e:
            logger.error(f"Error stopping server: {e}")
            raise

class EnhancedRouterGUI:
    """Enhanced router GUI with environment configuration support"""
    def __init__(self, root, router: EnhancedAPIRouter):
        self.root = root
        self.router = router
        self._fetch_result = None
        self.root.title("Enhanced API Router")
        self.setup_gui()
        
    def setup_gui(self):
        # Provider selection
        tk.Label(self.root, text="Provider:").grid(row=0, column=0, sticky='w', padx=5, pady=5)
        self.provider_var = tk.StringVar()
        self.provider_combo = ttk.Combobox(
            self.root,
            textvariable=self.provider_var,
            values=[p.value for p in ModelProvider]
        )
        self.provider_combo.grid(row=0, column=1, columnspan=2, sticky='ew', padx=5, pady=5)
        self.provider_combo.bind('<<ComboboxSelected>>', self.on_provider_selected)
        
        # Model selection
        tk.Label(self.root, text="Available Models:").grid(row=1, column=0, sticky='w', padx=5, pady=5)
        self.model_listbox = tk.Listbox(self.root, height=5, width=50)
        self.model_listbox.grid(row=2, column=0, columnspan=3, padx=5, pady=5, sticky='ew')
        
        # Control buttons
        self.fetch_btn = tk.Button(self.root, text="Fetch Models", command=self.fetch_models)
        self.fetch_btn.grid(row=3, column=0, padx=5, pady=5)
        
        self.start_btn = tk.Button(self.root, text="Start Server", command=self.start_server)
        self.start_btn.grid(row=3, column=1, padx=5, pady=5)
        
        self.stop_btn = tk.Button(
            self.root,
            text="Stop Server",
            command=self.stop_server,
            state='disabled'
        )
        self.stop_btn.grid(row=3, column=2, padx=5, pady=5)
        
        # Status display
        self.status_var = tk.StringVar(value="Stopped")
        tk.Label(self.root, textvariable=self.status_var).grid(
            row=4,
            column=0,
            columnspan=3,
            pady=10
        )

    def on_provider_selected(self, event=None):
        """Handle provider selection"""
        provider = ModelProvider(self.provider_var.get())
        api_key = self.router.get_api_key(provider)
        if not api_key:
            messagebox.showwarning(
                "Warning",
                f"No API key found for {provider.value} in .env file"
            )
            self.provider_combo.set('')

    async def _fetch_models(self):
        """Fetch models for selected provider"""
        provider = ModelProvider(self.provider_var.get())
        handler_class = self.router._handlers.get(provider)
        if handler_class:
            api_key = self.router.get_api_key(provider)
            if not api_key:
                raise ConfigurationError(f"No API key found for {provider.value}")
            handler = handler_class(api_key)
            return await handler.fetch_models()
        return []

    def fetch_models(self):
        if not self.provider_var.get():
            messagebox.showwarning("Warning", "Please select a provider")
            return
                    
        self.model_listbox.delete(0, tk.END)
        self.fetch_btn.configure(state='disabled')
        
        def fetch_and_update():
            try:
                models = asyncio.run(self._fetch_models())
                self.root.after(0, lambda: self._update_model_list(models))
            except Exception as e:
                self.root.after(0, lambda: messagebox.showerror("Error", str(e)))
            finally:
                self.root.after(0, lambda: self.fetch_btn.configure(state='normal'))
        
        threading.Thread(target=fetch_and_update, daemon=True).start()

    def _update_model_list(self, models):
        for model in models:
            self.model_listbox.insert(tk.END, f"{model.id} ({model.provider.value})")
        self.fetch_btn.configure(state='normal')


    def _check_fetch_result(self):
        try:
            models = self._fetch_result.result(timeout=0)
            for model in models:
                self.model_listbox.insert(tk.END, f"{model.id} ({model.provider.value})")
            self.fetch_btn.configure(state='normal')
        except concurrent.futures.TimeoutError:
            self.root.after(100, self._check_fetch_result)
        except Exception as e:
            messagebox.showerror("Error", str(e))
            self.fetch_btn.configure(state='normal')


    def start_server(self):
        if not self.model_listbox.curselection():
            messagebox.showwarning("Warning", "Please select a model first")
            return
                
        provider = ModelProvider(self.provider_var.get())
        selected_model = self.model_listbox.get(self.model_listbox.curselection())
        model_id = selected_model.split(" (")[0]  # Extract model ID from listbox display
        
        if self.router.start_server(provider):
            # Set the model name in the handler
            self.router.current_handler.model_name = model_id
            self.status_var.set(f"Running - Serving {provider.value} on http://localhost:{self.router.port}")
            self.update_button_states(running=True)


    def stop_server(self):
        if self.router.stop_server():
            self.status_var.set("Stopped")
            self.update_button_states(running=False)

    def update_button_states(self, running: bool):
        self.start_btn.configure(state='disabled' if running else 'normal')
        self.stop_btn.configure(state='normal' if running else 'disabled')
        self.provider_combo.configure(state='disabled' if running else 'normal')
        self.fetch_btn.configure(state='disabled' if running else 'normal')
        self.model_listbox.configure(state='disabled' if running else 'normal')

if __name__ == "__main__":
    root = tk.Tk()
    router = EnhancedAPIRouter()
    gui = EnhancedRouterGUI(root, router)
    root.mainloop()