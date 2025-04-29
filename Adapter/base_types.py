from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import List
from enum import Enum
from typing import Dict, List, Optional, Type, TypeVar, Generic, Any
import aiohttp
from contextlib import asynccontextmanager
import aiohttp
import logging
from abc import ABC, abstractmethod
from pydantic import BaseModel, Field, ValidationError

from functools import wraps
from datetime import datetime
import logging
T = TypeVar('T')
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

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - [%(threadName)s] - %(message)s'
)
logger = logging.getLogger(__name__)



class ModelProvider(Enum):
    ALIBABA = "Alibaba"
    DEEPINFRA = "DeepInfra"
    GITHUB_AZURE = "GitHub Azure"
    NVIDIA = "Nvidia"


class ProviderError(Exception):
    pass

class ModelFetchError(ProviderError):
    pass

class RequestProcessingError(ProviderError):
    pass

class ProviderConfig(BaseModel):
    """Provider configuration with validation"""
    base_url: str
    models_endpoint: Optional[str] = None
    auth_header: str = "Bearer"
    timeout: int = Field(default=30, ge=1, le=300)
    
    class Config:
        arbitrary_types_allowed = True



@dataclass
class ModelInfo:
    id: str
    provider: ModelProvider
    capabilities: List[str]

PROVIDER_CONFIGS: Dict[ModelProvider, ProviderConfig] = {
    ModelProvider.ALIBABA: ProviderConfig(
        base_url="https://dashscope-intl.aliyuncs.com/compatible-mode/v1",
        models_endpoint="/models"
    ),
    ModelProvider.DEEPINFRA: ProviderConfig(
        base_url="https://api.deepinfra.com/v1/openai",
        models_endpoint="/models"
    ),
    ModelProvider.GITHUB_AZURE: ProviderConfig(
        base_url="https://models.inference.ai.azure.com",
        auth_header="token"
    ),
    ModelProvider.NVIDIA: ProviderConfig(
        base_url="https://integrate.api.nvidia.com/v1",
        models_endpoint="/models"
    )
}


class BaseProviderHandler(ABC, Generic[T]):
    """Abstract base class for provider handlers"""
    def __init__(self, api_key: str):
        self.api_key = api_key
        self.config = PROVIDER_CONFIGS[self.provider_type]
        self._session: Optional[aiohttp.ClientSession] = None

    @property
    @abstractmethod
    def provider_type(self) -> ModelProvider:
        pass

    @asynccontextmanager
    async def get_session(self):
        """Async context manager for handling API sessions"""
        if self._session is None:
            async with aiohttp.ClientSession() as session:
                self._session = session
                try:
                    yield session
                finally:
                    self._session = None
        else:
            yield self._session

    @log_operation
    @abstractmethod
    async def fetch_models(self) -> List[ModelInfo]:
        pass

    @log_operation
    @abstractmethod
    async def process_request(self, path: str, request_data: dict) -> dict:
        pass
