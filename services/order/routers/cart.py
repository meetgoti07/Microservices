"""
Cart API Routes
"""
from fastapi import APIRouter, Depends, HTTPException, status
from typing import Dict, Any

from schemas import (
    AddToCartRequest, UpdateCartItemRequest,
    CartResponse, SuccessResponse
)
from services.cart_service import CartService
from database import get_database, get_redis
from jwt_middleware import get_current_user
import logging

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/cart", tags=["Cart"])


def get_cart_service():
    """Dependency to get cart service"""
    return CartService(get_database(), get_redis())


@router.get("", response_model=CartResponse)
async def get_cart(
    user: Dict[str, Any] = Depends(get_current_user),
    cart_service: CartService = Depends(get_cart_service)
):
    """Get user's shopping cart"""
    try:
        user_id = user.get("id")
        cart = await cart_service.get_cart(user_id)
        return cart
    except Exception as e:
        logger.error(f"Error getting cart: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve cart"
        )


@router.post("/items", response_model=CartResponse)
async def add_to_cart(
    request: AddToCartRequest,
    user: Dict[str, Any] = Depends(get_current_user),
    cart_service: CartService = Depends(get_cart_service)
):
    """Add item to cart"""
    try:
        user_id = user.get("id")
        cart = await cart_service.add_to_cart(user_id, request)
        return cart
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error adding to cart: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to add item to cart"
        )


@router.patch("/items/{cart_item_id}", response_model=CartResponse)
async def update_cart_item(
    cart_item_id: str,
    request: UpdateCartItemRequest,
    user: Dict[str, Any] = Depends(get_current_user),
    cart_service: CartService = Depends(get_cart_service)
):
    """Update cart item"""
    try:
        user_id = user.get("id")
        cart = await cart_service.update_cart_item(user_id, cart_item_id, request)
        return cart
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error updating cart item: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to update cart item"
        )


@router.delete("/items/{cart_item_id}", response_model=CartResponse)
async def remove_cart_item(
    cart_item_id: str,
    user: Dict[str, Any] = Depends(get_current_user),
    cart_service: CartService = Depends(get_cart_service)
):
    """Remove item from cart"""
    try:
        user_id = user.get("id")
        cart = await cart_service.remove_cart_item(user_id, cart_item_id)
        return cart
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error removing cart item: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to remove cart item"
        )


@router.delete("", response_model=SuccessResponse)
async def clear_cart(
    user: Dict[str, Any] = Depends(get_current_user),
    cart_service: CartService = Depends(get_cart_service)
):
    """Clear all items from cart"""
    try:
        user_id = user.get("id")
        await cart_service.clear_cart(user_id)
        return SuccessResponse(
            success=True,
            message="Cart cleared successfully"
        )
    except Exception as e:
        logger.error(f"Error clearing cart: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to clear cart"
        )
