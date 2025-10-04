"""
Utility functions for Order Service
"""
import random
import string
import json
from datetime import datetime, timedelta
from typing import Optional, Dict, Any
from decimal import Decimal
import httpx
from config import settings
import logging

logger = logging.getLogger(__name__)


def generate_order_number() -> str:
    """Generate unique order number"""
    timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
    random_suffix = ''.join(random.choices(string.digits, k=4))
    return f"{settings.ORDER_NUMBER_PREFIX}{timestamp}{random_suffix}"


def generate_order_token(token_number: int) -> tuple[str, str]:
    """
    Generate order token with prefix
    Returns: (full_token, prefix)
    """
    prefix = random.choice(settings.TOKEN_PREFIX_OPTIONS)
    token = f"{prefix}{token_number:03d}"
    return token, prefix


def calculate_tax(subtotal: float, tax_percentage: Optional[float] = None) -> float:
    """Calculate tax amount"""
    if tax_percentage is None:
        tax_percentage = settings.DEFAULT_TAX_PERCENTAGE
    
    tax = (subtotal * tax_percentage) / 100
    return round(tax, 2)


def calculate_total(subtotal: float, tax: float) -> float:
    """Calculate total amount"""
    return round(subtotal + tax, 2)


def estimate_preparation_time(item_count: int) -> int:
    """
    Estimate preparation time based on number of items
    Returns: time in minutes
    """
    base_time = 10  # 10 minutes base
    additional_time = item_count * 2  # 2 minutes per item
    return base_time + additional_time


def calculate_estimated_ready_time(preparation_time: int) -> datetime:
    """Calculate estimated ready time"""
    return datetime.utcnow() + timedelta(minutes=preparation_time)


async def fetch_menu_item_details(menu_item_id: str) -> Optional[Dict[str, Any]]:
    """
    Fetch menu item details from Menu Service
    """
    try:
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{settings.MENU_SERVICE_URL}/api/menu/items/{menu_item_id}",
                timeout=5.0
            )
            response.raise_for_status()
            result = response.json()
            
            # Menu service returns { success, message, data }
            # Extract the actual menu item from data field
            if isinstance(result, dict) and "data" in result:
                return result["data"]
            return result
    except Exception as e:
        logger.error(f"Error fetching menu item {menu_item_id}: {e}")
        return None


async def fetch_user_details(user_id: str, auth_token: str) -> Optional[Dict[str, Any]]:
    """
    Fetch user details from Auth Service
    """
    try:
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{settings.AUTH_SERVICE_URL}/api/user/{user_id}",
                headers={"Authorization": f"Bearer {auth_token}"},
                timeout=5.0
            )
            response.raise_for_status()
            return response.json()
    except Exception as e:
        logger.error(f"Error fetching user {user_id}: {e}")
        return None


async def notify_queue_service(
    order_id: str,
    user_id: str,
    auth_token: str,
    user_name: str = "",
    user_phone: str = "",
    item_count: int = 0
):
    """
    Notify Queue Service to create queue entry for order
    """
    try:
        async with httpx.AsyncClient() as client:
            await client.post(
                f"{settings.QUEUE_SERVICE_URL}/api/queue",
                json={
                    "order_id": order_id,
                    "user_id": user_id,
                    "user_name": user_name,
                    "user_phone": user_phone,
                    "item_count": item_count,
                    "priority": "NORMAL"
                },
                headers={
                    "Authorization": f"Bearer {auth_token}"
                },
                timeout=5.0
            )
    except Exception as e:
        logger.error(f"Error notifying queue service: {e}")


def serialize_datetime(obj):
    """JSON serializer for datetime objects"""
    if isinstance(obj, datetime):
        return obj.isoformat()
    raise TypeError(f"Type {type(obj)} not serializable")


def parse_objectid(obj: Any) -> Any:
    """Convert MongoDB ObjectId to string for JSON serialization"""
    if isinstance(obj, dict):
        return {k: parse_objectid(v) for k, v in obj.items()}
    elif isinstance(obj, list):
        return [parse_objectid(item) for item in obj]
    elif hasattr(obj, '__dict__'):
        return parse_objectid(obj.__dict__)
    else:
        return str(obj) if hasattr(obj, '__str__') and type(obj).__name__ == 'ObjectId' else obj


async def cache_get(redis_client, key: str) -> Optional[Any]:
    """Get value from Redis cache"""
    try:
        value = await redis_client.get(key)
        if value:
            return json.loads(value)
        return None
    except Exception as e:
        logger.error(f"Redis GET error: {e}")
        return None


async def cache_set(redis_client, key: str, value: Any, ttl: int = None):
    """Set value in Redis cache"""
    try:
        if ttl is None:
            ttl = settings.REDIS_CACHE_TTL
        
        await redis_client.setex(
            key,
            ttl,
            json.dumps(value, default=serialize_datetime)
        )
    except Exception as e:
        logger.error(f"Redis SET error: {e}")


async def cache_delete(redis_client, key: str):
    """Delete value from Redis cache"""
    try:
        await redis_client.delete(key)
    except Exception as e:
        logger.error(f"Redis DELETE error: {e}")


async def cache_delete_pattern(redis_client, pattern: str):
    """Delete all keys matching pattern"""
    try:
        keys = []
        async for key in redis_client.scan_iter(match=pattern):
            keys.append(key)
        
        if keys:
            await redis_client.delete(*keys)
    except Exception as e:
        logger.error(f"Redis DELETE pattern error: {e}")
