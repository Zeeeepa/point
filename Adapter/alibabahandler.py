from typing import List, Any
import logging
import asyncio
from abc import ABC, abstractmethod
from base_types import (
    BaseProviderHandler, 
    ModelInfo, 
    ModelProvider, 
    ModelFetchError, 
    RequestProcessingError
)

logger = logging.getLogger(__name__)

class AlibabaHandler(BaseProviderHandler):
    def __init__(self, api_key: str):
        super().__init__(api_key)
        self.model_name = None

    @property
    def provider_type(self) -> ModelProvider:
        return ModelProvider.ALIBABA
    
    async def fetch_models(self) -> List[ModelInfo]:
        headers = {
            'Authorization': f'Bearer {self.api_key}',
            'Content-Type': 'application/json'
        }
        
        async with self.get_session() as session:
            try:
                async with session.get(
                    f"{self.config.base_url}{self.config.models_endpoint}",
                    headers=headers,
                    timeout=self.config.timeout
                ) as response:
                    if response.status == 200:
                        data = await response.json()
                        return [
                            ModelInfo(
                                id=model['id'],
                                provider=self.provider_type,
                                capabilities=['chat', 'completion']
                            )
                            for model in data.get('data', [])
                        ]
                    raise ModelFetchError(f"Failed to fetch models: {response.status}")
            except Exception as e:
                logger.error(f"Error fetching Alibaba models: {e}")
                raise ModelFetchError(f"Failed to fetch models: {str(e)}")

    async def process_request(self, path: str, request_data: dict) -> dict:
        # Ensure model is set in the request data
        request_data['model'] = self.model_name
        
        headers = {
            'Authorization': f'Bearer {self.api_key}',
            'Content-Type': 'application/json'
        }
        
        async with self.get_session() as session:
            try:
                async with session.post(
                    f"{self.config.base_url}{path}",
                    json=request_data,
                    headers=headers,
                    timeout=self.config.timeout
                ) as response:
                    response_data = await response.json()
                    if response.status == 200:
                        return response_data
                    logger.error(f"Request failed with response: {response_data}")
                    raise RequestProcessingError(
                        f"Request failed with status {response.status}: {response_data}"
                    )
            except Exception as e:
                logger.error(f"Error processing request: {e}")
                raise RequestProcessingError(f"Request processing failed: {str(e)}")
