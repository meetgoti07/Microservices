"""
Configuration file for Order Service
"""
import os
from typing import Optional
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    # Service
    SERVICE_NAME: str = "order-service"
    SERVICE_PORT: int = 3003
    
    # MongoDB
    MONGODB_URL: str = os.getenv("MONGODB_URL", "mongodb://localhost:27017")
    MONGODB_DB_NAME: str = os.getenv("MONGODB_DB_NAME", "order_service")
    
    # Redis
    REDIS_URL: str = os.getenv("REDIS_URL", "redis://localhost:6379")
    REDIS_CACHE_TTL: int = 300  # 5 minutes
    
    # Auth Service
    AUTH_SERVICE_URL: str = os.getenv("AUTH_SERVICE_URL", "http://localhost:3001")
    
    # Menu Service
    MENU_SERVICE_URL: str = os.getenv("MENU_SERVICE_URL", "http://localhost:3002")
    
    # Queue Service
    QUEUE_SERVICE_URL: str = os.getenv("QUEUE_SERVICE_URL", "http://localhost:3004")
    
    # Kafka (Optional)
    KAFKA_BROKERS: Optional[str] = os.getenv("KAFKA_BROKERS", "localhost:9092")
    
    # Python Environment
    PYTHONUNBUFFERED: Optional[str] = os.getenv("PYTHONUNBUFFERED", "1")
    
    # Tax Configuration
    DEFAULT_TAX_PERCENTAGE: float = 5.0  # 5% tax
    
    # Order Configuration
    ORDER_NUMBER_PREFIX: str = "ORD"
    TOKEN_PREFIX_OPTIONS: list = ["A", "B", "C", "D", "E"]
    MAX_CART_ITEMS: int = 50
    MAX_QUANTITY_PER_ITEM: int = 50
    
    # Pagination
    DEFAULT_PAGE_SIZE: int = 20
    MAX_PAGE_SIZE: int = 100
    
    class Config:
        env_file = ".env"
        case_sensitive = True
        extra = "ignore"  # Ignore extra fields from .env


settings = Settings()
