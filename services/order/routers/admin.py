"""
Admin/Staff API Routes
"""
from fastapi import APIRouter, Depends, HTTPException, status, Query
from typing import Dict, Any, Optional, List
from datetime import datetime, date

from schemas import (
    OrderResponse, OrderListResponse, OrderStatus,
    PaginatedResponse, OrderStatisticsResponse
)
from services.order_service import OrderService
from database import get_database, get_redis
from jwt_middleware import get_current_user
import logging

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/admin", tags=["Admin"])


def get_order_service():
    """Dependency to get order service"""
    return OrderService(get_database(), get_redis())


def require_staff(user: Dict[str, Any]) -> Dict[str, Any]:
    """Dependency to require staff role"""
    role = user.get("role", "")
    roles = user.get("roles", [])
    
    # Check both single role string and roles array for compatibility
    if role not in ["staff", "admin"] and "staff" not in roles and "admin" not in roles:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Staff access required"
        )
    return user


@router.get("/orders", response_model=PaginatedResponse)
async def get_all_orders(
    status_filter: Optional[OrderStatus] = Query(None, alias="status"),
    search: Optional[str] = Query(None),
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    user: Dict[str, Any] = Depends(get_current_user),
    order_service: OrderService = Depends(get_order_service)
):
    """Get all orders (Staff only)"""
    require_staff(user)
    
    try:
        db = get_database()
        
        # Build query
        query = {}
        if status_filter:
            query["status"] = status_filter
        
        if search:
            query["$or"] = [
                {"order_number": {"$regex": search, "$options": "i"}},
                {"user_name": {"$regex": search, "$options": "i"}},
                {"user_email": {"$regex": search, "$options": "i"}}
            ]
        
        # Get total count
        total = await db.orders.count_documents(query)
        
        # Get orders
        skip = (page - 1) * page_size
        cursor = db.orders.find(query).sort("created_at", -1).skip(skip).limit(page_size)
        
        orders = []
        async for order in cursor:
            # Get token
            token_data = await db.order_tokens.find_one({"order_id": order["id"]})
            token = token_data["token"] if token_data else None
            
            # Count items
            item_count = await db.order_items.count_documents({"order_id": order["id"]})
            
            orders.append(OrderListResponse(
                id=order["id"],
                order_number=order["order_number"],
                status=order["status"],
                total=order["total"],
                item_count=item_count,
                estimated_ready_time=order.get("estimated_ready_time"),
                token=token,
                created_at=order["created_at"]
            ))
        
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
        logger.error(f"Error getting all orders: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve orders"
        )


@router.get("/orders/active", response_model=List[OrderListResponse])
async def get_active_orders(
    user: Dict[str, Any] = Depends(get_current_user),
    order_service: OrderService = Depends(get_order_service)
):
    """Get all active orders (Staff only)"""
    require_staff(user)
    
    try:
        db = get_database()
        
        # Query for active orders (not completed, cancelled, or refunded)
        query = {
            "status": {
                "$in": [
                    OrderStatus.PENDING,
                    OrderStatus.CONFIRMED,
                    OrderStatus.PREPARING,
                    OrderStatus.READY
                ]
            }
        }
        
        cursor = db.orders.find(query).sort("estimated_ready_time", 1)
        
        orders = []
        async for order in cursor:
            # Get token
            token_data = await db.order_tokens.find_one({"order_id": order["id"]})
            token = token_data["token"] if token_data else None
            
            # Count items
            item_count = await db.order_items.count_documents({"order_id": order["id"]})
            
            orders.append(OrderListResponse(
                id=order["id"],
                order_number=order["order_number"],
                status=order["status"],
                total=order["total"],
                item_count=item_count,
                estimated_ready_time=order.get("estimated_ready_time"),
                token=token,
                created_at=order["created_at"]
            ))
        
        return orders
    except Exception as e:
        logger.error(f"Error getting active orders: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve active orders"
        )


@router.get("/statistics/today", response_model=Dict[str, Any])
async def get_today_statistics(
    user: Dict[str, Any] = Depends(get_current_user)
):
    """Get today's order statistics (Staff only)"""
    require_staff(user)
    
    try:
        db = get_database()
        
        # Get today's date range
        today = datetime.utcnow().date()
        start_of_day = datetime.combine(today, datetime.min.time())
        end_of_day = datetime.combine(today, datetime.max.time())
        
        # Get today's orders
        query = {
            "created_at": {
                "$gte": start_of_day,
                "$lte": end_of_day
            }
        }
        
        cursor = db.orders.find(query)
        
        # Calculate statistics
        total_orders = 0
        status_counts = {status.value: 0 for status in OrderStatus}
        total_revenue = 0.0
        total_tax = 0.0
        preparation_times = []
        
        async for order in cursor:
            total_orders += 1
            status_counts[order["status"]] += 1
            
            if order["status"] not in [OrderStatus.CANCELLED, OrderStatus.REFUNDED]:
                total_revenue += order["total"]
                total_tax += order["tax"]
            
            if order.get("actual_preparation_time"):
                preparation_times.append(order["actual_preparation_time"])
        
        avg_preparation_time = (
            sum(preparation_times) / len(preparation_times)
            if preparation_times else 0
        )
        
        avg_order_value = total_revenue / total_orders if total_orders > 0 else 0.0
        
        return {
            "date": today.isoformat(),
            "total_orders": total_orders,
            "pending_orders": status_counts[OrderStatus.PENDING],
            "confirmed_orders": status_counts[OrderStatus.CONFIRMED],
            "preparing_orders": status_counts[OrderStatus.PREPARING],
            "ready_orders": status_counts[OrderStatus.READY],
            "completed_orders": status_counts[OrderStatus.COMPLETED],
            "cancelled_orders": status_counts[OrderStatus.CANCELLED],
            "refunded_orders": status_counts[OrderStatus.REFUNDED],
            "total_revenue": round(total_revenue, 2),
            "total_tax": round(total_tax, 2),
            "avg_order_value": round(avg_order_value, 2),
            "avg_preparation_time": int(avg_preparation_time)
        }
    except Exception as e:
        logger.error(f"Error getting statistics: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve statistics"
        )


@router.get("/statistics/range", response_model=List[OrderStatisticsResponse])
async def get_statistics_range(
    start_date: date = Query(...),
    end_date: date = Query(...),
    user: Dict[str, Any] = Depends(get_current_user)
):
    """Get order statistics for date range (Staff only)"""
    require_staff(user)
    
    try:
        db = get_database()
        
        query = {
            "date": {
                "$gte": start_date.isoformat(),
                "$lte": end_date.isoformat()
            }
        }
        
        cursor = db.order_statistics.find(query).sort("date", 1)
        
        statistics = []
        async for stat in cursor:
            statistics.append(OrderStatisticsResponse(**stat))
        
        return statistics
    except Exception as e:
        logger.error(f"Error getting statistics range: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve statistics"
        )


@router.get("/payments", response_model=PaginatedResponse)
async def get_all_payments(
    status_filter: Optional[str] = Query(None, alias="status"),
    method_filter: Optional[str] = Query(None, alias="method"),
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    user: Dict[str, Any] = Depends(get_current_user)
):
    """Get all payments (Admin only)"""
    roles = user.get("roles", [])
    if "admin" not in roles:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Admin access required"
        )
    
    try:
        db = get_database()
        
        # Build query
        query = {}
        if status_filter:
            query["status"] = status_filter
        if method_filter:
            query["method"] = method_filter
        
        # Get total count
        total = await db.payments.count_documents(query)
        
        # Get payments
        skip = (page - 1) * page_size
        cursor = db.payments.find(query).sort("initiated_at", -1).skip(skip).limit(page_size)
        
        payments = []
        async for payment in cursor:
            # Get order info
            order = await db.orders.find_one({"id": payment["order_id"]})
            payment["order_number"] = order["order_number"] if order else None
            payments.append(payment)
        
        total_pages = (total + page_size - 1) // page_size
        
        return PaginatedResponse(
            items=payments,
            total=total,
            page=page,
            page_size=page_size,
            total_pages=total_pages,
            has_next=page < total_pages,
            has_prev=page > 1
        )
    except Exception as e:
        logger.error(f"Error getting payments: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve payments"
        )
