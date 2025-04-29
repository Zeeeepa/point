from typing import List
import logging
import asyncio
from azure.ai.inference import ChatCompletionsClient
from azure.ai.inference.models import AssistantMessage, SystemMessage, UserMessage
from azure.core.credentials import AzureKeyCredential
from base_types import (
    BaseProviderHandler,
    ModelInfo,
    ModelProvider,
    ModelFetchError,
    RequestProcessingError
)

logger = logging.getLogger(__name__)

class GithubAzureHandler(BaseProviderHandler):
    """Handler for GitHub DeepSeek API provider using Azure AI SDK"""
    
    def __init__(self, api_key: str):
        super().__init__(api_key)
        self.client = ChatCompletionsClient(
            endpoint=self.config.base_url,
            credential=AzureKeyCredential(api_key)
        )
        self.model_name = "DeepSeek-R1"  # Default model

    @property
    def provider_type(self) -> ModelProvider:
        return ModelProvider.GITHUB_DEEPSEEK

    async def fetch_models(self) -> List[ModelInfo]:
        """Fetch available DeepSeek models"""
        try:
            # Azure SDK doesn't provide model listing, return default model
            return [
                ModelInfo(
                    id=self.model_name,
                    provider=self.provider_type,
                    capabilities=['chat', 'completion', 'code']
                )
            ]
        except Exception as e:
            logger.error(f"Error fetching GitHub DeepSeek models: {e}")
            raise ModelFetchError(f"Failed to fetch models: {str(e)}")

    async def process_request(self, path: str, request_data: dict) -> dict:
        """Process requests using Azure AI SDK"""
        try:
            if path == '/chat/completions':
                # Convert request format to Azure SDK messages
                messages = []
                for msg in request_data['messages']:
                    if msg['role'] == 'system':
                        messages.append(SystemMessage(content=msg['content']))
                    elif msg['role'] == 'assistant':
                        messages.append(AssistantMessage(content=msg['content']))
                    else:  # user messages
                        messages.append(UserMessage(content=msg['content']))

                # Make request through Azure SDK
                response = await asyncio.to_thread(
                    lambda: self.client.complete(
                        messages=messages,
                        model=request_data.get('model', self.model_name)
                    )
                )

                # Convert response to expected format
                return {
                    'choices': [{
                        'message': {
                            'content': response.choices[0].message.content,
                            'role': 'assistant'
                        }
                    }],
                    'usage': {
                        'total_tokens': response.usage.total_tokens if response.usage else 0
                    }
                }
            else:
                raise RequestProcessingError(f"Unsupported endpoint: {path}")

        except Exception as e:
            logger.error(f"Error processing GitHub DeepSeek request: {e}")
            raise RequestProcessingError(f"Request failed: {str(e)}")
        
