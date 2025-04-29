from typing import List, Any
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

class DeepInfraHandler(BaseProviderHandler):
    """Handler for DeepInfra API provider using OpenAI-compatible interface"""
    
    def __init__(self, api_key: str):
        super().__init__(api_key)
        self.client = OpenAI(
            api_key=api_key,
            base_url=self.config.base_url
        )

    @property
    def provider_type(self) -> ModelProvider:
        return ModelProvider.DEEPINFRA

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
            logger.error(f"Error fetching DeepInfra models: {e}")
            raise ModelFetchError(f"Failed to fetch models: {str(e)}")

    def _determine_capabilities(self, model_data: Any) -> List[str]:
        """Determine model capabilities from OpenAI model data"""
        capabilities = ['chat', 'completion']  # Base capabilities
        
        # Add additional capabilities based on model properties
        model_id = str(model_data.id).lower()
        if 'instruct' in model_id:
            capabilities.append('instruction')
        if 'code' in model_id:
            capabilities.append('code')
            
        return capabilities

    async def process_request(self, path: str, request_data: dict) -> dict:
        """Process requests using OpenAI-compatible endpoints"""
        try:
            # Convert path to endpoint type
            if path == '/chat/completions':
                response = await asyncio.to_thread(
                    lambda: self.client.chat.completions.create(
                        **request_data
                    )
                )
                # Convert response to dict format expected by router
                return {
                    'choices': [{
                        'message': {
                            'content': choice.message.content,
                            'role': choice.message.role
                        }
                    } for choice in response.choices],
                    'usage': {
                        'prompt_tokens': response.usage.prompt_tokens,
                        'completion_tokens': response.usage.completion_tokens,
                        'total_tokens': response.usage.total_tokens
                    }
                }
            else:
                raise RequestProcessingError(f"Unsupported endpoint: {path}")
                
        except Exception as e:
            logger.error(f"Error processing DeepInfra request: {e}")
            raise RequestProcessingError(f"Request failed: {str(e)}")