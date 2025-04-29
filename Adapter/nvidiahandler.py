from typing import List, Any, AsyncGenerator, Dict, Union
import logging
import asyncio
from openai import OpenAI
from base_types import (
    BaseProviderHandler,
    ModelInfo,
    ModelProvider,
    ModelFetchError,
    RequestProcessingError
)

logger = logging.getLogger(__name__)

class NvidiaHandler(BaseProviderHandler):
    """Handler for Nvidia API provider using OpenAI-compatible interface"""
    
    def __init__(self, api_key: str):
        super().__init__(api_key)
        # Initialize client with the configuration from base_types
        self.client = OpenAI(
            api_key=api_key,
            base_url=self.config.base_url
        )
        self.model_name = None  # Will be set by the router

    @property
    def provider_type(self) -> ModelProvider:
        return ModelProvider.NVIDIA

    async def fetch_models(self) -> List[ModelInfo]:
        """Fetch available models using OpenAI-compatible endpoint"""
        try:
            # OpenAI client is synchronous, wrap in asyncio
            models = await asyncio.to_thread(
                lambda: self.client.models.list()
            )
            return [
                ModelInfo(
                    id=str(model.id),
                    provider=self.provider_type,
                    capabilities=self._determine_capabilities(model)
                )
                for model in models.data
            ]
        except Exception as e:
            logger.error(f"Error fetching Nvidia models: {e}")
            raise ModelFetchError(f"Failed to fetch models: {str(e)}")

    def _determine_capabilities(self, model_data: Any) -> List[str]:
        """Determine model capabilities from OpenAI model data"""
        capabilities = ['chat', 'completion']  # Base capabilities
        
        # Add additional capabilities based on model properties
        model_id = str(model_data.id).lower()
        if 'instruct' in model_id:
            capabilities.append('instruction')
        if 'code' in model_id or 'deepseek' in model_id:
            capabilities.append('code')
        if any(name in model_id for name in ['vision', 'claude-3', 'gpt-4v']):
            capabilities.append('vision')
            
        return capabilities

    async def process_request(self, path: str, request_data: dict) -> Union[dict, AsyncGenerator[dict, None]]:
        """Process requests using OpenAI-compatible endpoints"""
        try:
            # Handle streaming parameter
            is_streaming = request_data.get('stream', False)
            
            # Ensure model is set in the request data
            if self.model_name and 'model' not in request_data:
                request_data['model'] = self.model_name
            
            # Fix message content format if needed
            if path == '/chat/completions' and 'messages' in request_data:
                self._normalize_messages(request_data)
            
            # Convert path to endpoint type
            if path == '/chat/completions':
                if is_streaming:
                    # For streaming, we need to handle the response differently
                    return await self._stream_chat_completion(request_data)
                else:
                    return await self._handle_chat_completion(request_data)
            elif path == '/completions':
                if is_streaming:
                    return await self._stream_text_completion(request_data)
                else:
                    return await self._handle_text_completion(request_data)
            else:
                raise RequestProcessingError(f"Unsupported endpoint: {path}")
                
        except Exception as e:
            logger.error(f"Error processing Nvidia request: {e}")
            raise RequestProcessingError(f"Request failed: {str(e)}")
    
    def _normalize_messages(self, request_data: dict):
        """Normalize message format for compatibility with NVIDIA API"""
        for i, msg in enumerate(request_data['messages']):
            # Check if content is a list (multimodal format)
            if isinstance(msg.get('content'), list):
                # For text-only models, convert to simple text
                if not self._model_supports_vision(request_data['model']):
                    text_parts = []
                    for part in msg['content']:
                        if part.get('type') == 'text':
                            text_parts.append(part.get('text', ''))
                    request_data['messages'][i]['content'] = ' '.join(text_parts)
    
    def _model_supports_vision(self, model_id: str) -> bool:
        """Check if model supports vision/multimodal inputs"""
        vision_indicators = ['vision', 'claude-3', 'gpt-4v', 'llava']
        return any(indicator in model_id.lower() for indicator in vision_indicators)
    
    async def _handle_chat_completion(self, request_data: dict) -> dict:
        """Handle regular (non-streaming) chat completions"""
        response = await asyncio.to_thread(
            lambda: self.client.chat.completions.create(
                **request_data
            )
        )
        
        # Convert response to dict format expected by router
        return {
            'id': response.id,
            'object': response.object,
            'created': response.created,
            'model': response.model,
            'choices': [{
                'index': choice.index,
                'message': {
                    'content': choice.message.content,
                    'role': choice.message.role
                },
                'finish_reason': choice.finish_reason
            } for choice in response.choices],
            'usage': {
                'prompt_tokens': response.usage.prompt_tokens,
                'completion_tokens': response.usage.completion_tokens,
                'total_tokens': response.usage.total_tokens
            }
        }
    
    async def _stream_chat_completion(self, request_data: dict) -> AsyncGenerator[dict, None]:
        """Stream chat completion responses"""
        # Start the streaming request using the client
        stream = await asyncio.to_thread(
            lambda: self.client.chat.completions.create(
                **request_data
            )
        )
        
        # Yield chunks as they arrive
        for chunk in stream:
            # Extract content from the chunk
            delta_content = chunk.choices[0].delta.content
            if delta_content is not None:
                # Yield the chunk directly for streaming
                yield {
                    'id': chunk.id,
                    'object': 'chat.completion.chunk',
                    'created': chunk.created,
                    'model': chunk.model,
                    'choices': [{
                        'index': chunk.choices[0].index,
                        'delta': {'content': delta_content},
                        'finish_reason': chunk.choices[0].finish_reason
                    }]
                }
    
    async def _handle_text_completion(self, request_data: dict) -> dict:
        """Handle regular (non-streaming) text completions"""
        response = await asyncio.to_thread(
            lambda: self.client.completions.create(
                **request_data
            )
        )
        
        # Handle completions response format
        return {
            'id': response.id,
            'object': response.object,
            'created': response.created,
            'model': response.model,
            'choices': [{
                'text': choice.text,
                'index': choice.index,
                'finish_reason': choice.finish_reason
            } for choice in response.choices],
            'usage': {
                'prompt_tokens': response.usage.prompt_tokens,
                'completion_tokens': response.usage.completion_tokens,
                'total_tokens': response.usage.total_tokens
            }
        }
    
    async def _stream_text_completion(self, request_data: dict) -> AsyncGenerator[dict, None]:
        """Stream text completion responses"""
        # Start the streaming request
        stream = await asyncio.to_thread(
            lambda: self.client.completions.create(
                **request_data
            )
        )
        
        # Yield chunks as they arrive
        for chunk in stream:
            # Extract text from the chunk
            chunk_text = chunk.choices[0].text
            if chunk_text is not None:
                # Yield the chunk directly for streaming
                yield {
                    'id': chunk.id,
                    'object': 'text_completion.chunk',
                    'created': chunk.created,
                    'model': chunk.model,
                    'choices': [{
                        'text': chunk_text,
                        'index': chunk.choices[0].index,
                        'finish_reason': chunk.choices[0].finish_reason
                    }]
                }
