"""
Order API Routes
"""
from fastapi import APIRouter, Depends, HTTPException, status, Query, Header
from typing import Dict, Any, Optional, Annotated

from schemas import (
    CreateOrderRequest, UpdateOrderStatusRequest, CancelOrderRequest,
    OrderFeedbackRequest, OrderResponse, OrderListResponse,
    OrderStatus, SuccessResponse, PaginatedResponse
)
from services.order_service import OrderService
from database import get_database, get_redis
from jwt_middleware import get_current_user
from config import settings
import logging

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/orders", tags=["Orders"])


def get_order_service():
    """Dependency to get order service"""
    return OrderService(get_database(), get_redis())


def check_staff_role(user: Dict[str, Any]) -> bool:
    """Check if user has staff role"""
    role = user.get("role", "")
    roles = user.get("roles", [])
    # Check both single role string and roles array for compatibility
    return role in ["staff", "admin"] or "staff" in roles or "admin" in roles


@router.post("", response_model=OrderResponse, status_code=status.HTTP_201_CREATED)
async def create_order(
    request: CreateOrderRequest,
    authorization: Annotated[Optional[str], Header()] = None,
    user: Dict[str, Any] = Depends(get_current_user),
    order_service: OrderService = Depends(get_order_service)
):
    """Create order from cart"""
    try:
        user_id = user.get("id")
        # Extract token from Authorization header
        token = authorization.replace("Bearer ", "") if authorization else None
        order = await order_service.create_order(user_id, user, request, token)
        return order
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error creating order: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to create order"
        )


@router.get("/my-orders", response_model=PaginatedResponse)
async def get_my_orders(
    status_filter: Optional[OrderStatus] = Query(None, alias="status"),
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    user: Dict[str, Any] = Depends(get_current_user),
    order_service: OrderService = Depends(get_order_service)
):
    """Get user's orders with pagination"""
    try:
        user_id = user.get("id")
        orders, total = await order_service.get_my_orders(
            user_id,
            status=status_filter,
            page=page,
            page_size=page_size
        )
        
        total_pages = (total + page_size - 1) // page_size
        
        return PaginatedResponse(
            items=[order.model_dump() for order in orders],
            total=total,
            page=page,
            page_size=page_size,
            total_pages=total_pages,
            has_next=page < total_pages,
            has_prev=page > 1
        )
    except Exception as e:
        logger.error(f"Error getting orders: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve orders"
        )


@router.get("/{order_id}", response_model=OrderResponse)
async def get_order(
    order_id: str,
    user: Dict[str, Any] = Depends(get_current_user),
    order_service: OrderService = Depends(get_order_service)
):
    """Get order details"""
    try:
        user_id = user.get("id")
        is_staff = check_staff_role(user)
        
        # Staff can view any order, users can only view their own
        order = await order_service.get_order(
            order_id,
            user_id=None if is_staff else user_id
        )
        return order
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error getting order: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve order"
        )


@router.get("/{order_id}/status", response_model=Dict[str, Any])
async def get_order_status(
    order_id: str,
    order_service: OrderService = Depends(get_order_service)
):
    """Get order status (public endpoint)"""
    try:
        order = await order_service.get_order(order_id)
        return {
            "order_id": order.id,
            "order_number": order.order_number,
            "status": order.status,
            "token": order.token.token if order.token else None,
            "estimated_ready_time": order.estimated_ready_time,
            "actual_ready_time": order.actual_ready_time
        }
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error getting order status: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve order status"
        )


@router.patch("/{order_id}/status", response_model=OrderResponse)
async def update_order_status(
    order_id: str,
    request: UpdateOrderStatusRequest,
    user: Dict[str, Any] = Depends(get_current_user),
    order_service: OrderService = Depends(get_order_service)
):
    """Update order status (Staff only)"""
    try:
        is_staff = check_staff_role(user)
        if not is_staff:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Only staff can update order status"
            )
        
        user_id = user.get("id")
        user_name = user.get("name", "Staff")
        
        order = await order_service.update_order_status(
            order_id,
            request,
            user_id,
            user_name,
            is_staff=True
        )
        return order
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error updating order status: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to update order status"
        )


@router.post("/{order_id}/cancel", response_model=OrderResponse)
async def cancel_order(
    order_id: str,
    request: CancelOrderRequest,
    user: Dict[str, Any] = Depends(get_current_user),
    order_service: OrderService = Depends(get_order_service)
):
    """Cancel order"""
    try:
        user_id = user.get("id")
        order = await order_service.cancel_order(order_id, user_id, request)
        return order
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error cancelling order: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to cancel order"
        )


@router.post("/{order_id}/feedback", response_model=OrderResponse)
async def submit_feedback(
    order_id: str,
    request: OrderFeedbackRequest,
    user: Dict[str, Any] = Depends(get_current_user),
    order_service: OrderService = Depends(get_order_service)
):
    """Submit order feedback"""
    try:
        user_id = user.get("id")
        order = await order_service.submit_feedback(order_id, user_id, request)
        return order
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    except Exception as e:
        logger.error(f"Error submitting feedback: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to submit feedback"
        )
