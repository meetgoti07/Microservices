"""
Database connection and setup for MongoDB and Redis
"""
from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorDatabase
from redis.asyncio import Redis
from typing import Optional
from config import settings
import logging

logger = logging.getLogger(__name__)


class Database:
    client: Optional[AsyncIOMotorClient] = None
    db: Optional[AsyncIOMotorDatabase] = None
    redis: Optional[Redis] = None


db = Database()


async def connect_to_mongo():
    """Connect to MongoDB"""
    try:
        db.client = AsyncIOMotorClient(settings.MONGODB_URL)
        db.db = db.client[settings.MONGODB_DB_NAME]
        
        # Test connection
        await db.client.admin.command('ping')
        logger.info(f"Connected to MongoDB: {settings.MONGODB_DB_NAME}")
        
        # Create indexes
        await create_indexes()
        
    except Exception as e:
        logger.error(f"Could not connect to MongoDB: {e}")
        raise


async def close_mongo_connection():
    """Close MongoDB connection"""
    if db.client:
        db.client.close()
        logger.info("Closed MongoDB connection")


async def connect_to_redis():
    """Connect to Redis"""
    try:
        db.redis = Redis.from_url(
            settings.REDIS_URL,
            encoding="utf-8",
            decode_responses=True
        )
        
        # Test connection
        await db.redis.ping()
        logger.info("Connected to Redis")
        
    except Exception as e:
        logger.error(f"Could not connect to Redis: {e}")
        raise


async def close_redis_connection():
    """Close Redis connection"""
    if db.redis:
        await db.redis.close()
        logger.info("Closed Redis connection")


async def create_indexes():
    """Create database indexes for better performance"""
    try:
        # Shopping Carts indexes
        await db.db.shopping_carts.create_index("user_id", unique=True)
        await db.db.shopping_carts.create_index("updated_at")
        
        # Cart Items indexes
        await db.db.cart_items.create_index("cart_id")
        await db.db.cart_items.create_index("menu_item_id")
        await db.db.cart_items.create_index("added_at")
        
        # Orders indexes
        await db.db.orders.create_index("order_number", unique=True)
        await db.db.orders.create_index("user_id")
        await db.db.orders.create_index("status")
        await db.db.orders.create_index("created_at")
        await db.db.orders.create_index("estimated_ready_time")
        await db.db.orders.create_index([("user_id", 1), ("created_at", -1)])
        
        # Order Tokens indexes
        await db.db.order_tokens.create_index("order_id", unique=True)
        await db.db.order_tokens.create_index("token", unique=True)
        await db.db.order_tokens.create_index("generated_at")
        
        # Order Items indexes
        await db.db.order_items.create_index("order_id")
        await db.db.order_items.create_index("menu_item_id")
        
        # Order Timeline indexes
        await db.db.order_timeline.create_index("order_id")
        await db.db.order_timeline.create_index("timestamp")
        
        # Payments indexes
        await db.db.payments.create_index("order_id", unique=True)
        await db.db.payments.create_index("status")
        await db.db.payments.create_index("transaction_id")
        await db.db.payments.create_index("initiated_at")
        
        # Order Statistics indexes
        await db.db.order_statistics.create_index("date", unique=True)
        
        logger.info("Database indexes created successfully")
        
    except Exception as e:
        logger.error(f"Error creating indexes: {e}")


def get_database() -> AsyncIOMotorDatabase:
    """Get database instance"""
    return db.db


def get_redis() -> Redis:
    """Get Redis instance"""
    return db.redis
