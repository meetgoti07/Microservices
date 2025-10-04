"""
Order Service - FastAPI Application
Manages orders, shopping carts, and payments with MongoDB and Redis
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import logging

from database import (
    connect_to_mongo, close_mongo_connection,
    connect_to_redis, close_redis_connection
)
from routers import cart, orders, admin
from config import settings

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events"""
    # Startup
    logger.info(f"Starting {settings.SERVICE_NAME}...")
    try:
        await connect_to_mongo()
        await connect_to_redis()
        logger.info("Database connections established")
    except Exception as e:
        logger.error(f"Failed to connect to databases: {e}")
        raise
    
    yield
    
    # Shutdown
    logger.info(f"Shutting down {settings.SERVICE_NAME}...")
    await close_mongo_connection()
    await close_redis_connection()
    logger.info("Database connections closed")


# Create FastAPI app
app = FastAPI(
    title="Order Service API",
    description="Microservice for managing orders, shopping carts, and payments",
    version="1.0.0",
    lifespan=lifespan
)

# Configure CORS
# CORS is handled by nginx, but we keep it for direct access during development
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:3000",
        "http://localhost:3001",
        "http://localhost:3002",
        "http://localhost:3003",
        "http://localhost:3004",
        "http://localhost:3005",
        "http://localhost:8080",
        "http://127.0.0.1:3000",
        "http://127.0.0.1:8080",
    ],
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"],
    allow_headers=["Content-Type", "Authorization"],
    expose_headers=["Content-Length", "Content-Type"],
)

# Add response headers middleware
@app.middleware("http")
async def add_cors_headers(request, call_next):
    """Add CORS headers to all responses"""
    response = await call_next(request)
    
    # Get origin from request
    origin = request.headers.get("origin")
    
    # Check if origin is allowed
    allowed_origins = [
        "http://localhost:3000",
        "http://localhost:8080",
        "http://127.0.0.1:3000",
        "http://127.0.0.1:8080",
    ]
    
    if origin in allowed_origins:
        response.headers["Access-Control-Allow-Origin"] = origin
        response.headers["Access-Control-Allow-Credentials"] = "true"
        response.headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, PATCH, OPTIONS"
        response.headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"
    
    return response

# Include routers
app.include_router(cart.router)
app.include_router(orders.router)
app.include_router(admin.router)


@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "service": settings.SERVICE_NAME,
        "version": "1.0.0",
        "status": "running"
    }


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": settings.SERVICE_NAME
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.SERVICE_PORT,
        reload=True,
        log_level="info"
    )