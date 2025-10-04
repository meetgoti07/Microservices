"""
Order Service - Business logic for order operations
"""
from typing import Optional, List, Dict, Any
from datetime import datetime, timedelta
from motor.motor_asyncio import AsyncIOMotorDatabase
from redis.asyncio import Redis
import uuid

from schemas import (
    CreateOrderRequest, UpdateOrderStatusRequest, CancelOrderRequest,
    OrderFeedbackRequest, OrderItemFeedbackRequest,
    OrderResponse, OrderListResponse, OrderStatus, PaymentStatus,
    OrderItemResponse, OrderTokenResponse, PaymentResponse, OrderTimelineResponse
)
from utils import (
    generate_order_number, generate_order_token,
    estimate_preparation_time, calculate_estimated_ready_time,
    notify_queue_service, cache_get, cache_set, cache_delete, cache_delete_pattern
)
from config import settings
import logging

logger = logging.getLogger(__name__)


class OrderService:
    def __init__(self, db: AsyncIOMotorDatabase, redis: Redis):
        self.db = db
        self.redis = redis
    
    async def create_order(
        self, 
        user_id: str,
        user_data: Dict[str, Any],
        request: CreateOrderRequest,
        auth_token: Optional[str] = None
    ) -> OrderResponse:
        """Create order from cart"""
        # Get cart
        cart = await self.db.shopping_carts.find_one({"user_id": user_id})
        if not cart or cart["item_count"] == 0:
            raise ValueError("Cart is empty")
        
        # Get cart items
        cart_items = []
        async for item in self.db.cart_items.find({"cart_id": cart["id"]}):
            # Get customizations
            customizations = []
            async for custom in self.db.cart_item_customizations.find({"cart_item_id": item["id"]}):
                customizations.append(custom)
            item["customizations"] = customizations
            cart_items.append(item)
        
        if not cart_items:
            raise ValueError("Cart is empty")
        
        # Generate order details
        order_id = str(uuid.uuid4())
        order_number = generate_order_number()
        
        # Calculate preparation time
        prep_time = estimate_preparation_time(cart["item_count"])
        estimated_ready = calculate_estimated_ready_time(prep_time)
        
        # Create order
        order = {
            "id": order_id,
            "order_number": order_number,
            "user_id": user_id,
            "user_name": user_data.get("name"),
            "user_email": user_data.get("email"),
            "user_phone": user_data.get("phone"),
            "status": OrderStatus.PENDING,
            "subtotal": cart["subtotal"],
            "tax": cart["tax"],
            "total": cart["total"],
            "special_instructions": request.special_instructions,
            "estimated_preparation_time": prep_time,
            "estimated_ready_time": estimated_ready,
            "created_at": datetime.utcnow(),
            "updated_at": datetime.utcnow()
        }
        
        await self.db.orders.insert_one(order)
        
        # Create order items
        for cart_item in cart_items:
            order_item_id = str(uuid.uuid4())
            order_item = {
                "id": order_item_id,
                "order_id": order_id,
                "menu_item_id": cart_item["menu_item_id"],
                "menu_item_name": cart_item["menu_item_name"],
                "menu_item_image": cart_item.get("menu_item_image"),
                "quantity": cart_item["quantity"],
                "unit_price": cart_item["unit_price"],
                "total_price": cart_item["total_price"],
                "special_instructions": cart_item.get("special_instructions")
            }
            await self.db.order_items.insert_one(order_item)
            
            # Create order item customizations
            for custom in cart_item["customizations"]:
                customization = {
                    "id": str(uuid.uuid4()),
                    "order_item_id": order_item_id,
                    "customization_id": custom["customization_id"],
                    "customization_name": custom["customization_name"],
                    "selected_label": custom["selected_label"],
                    "selected_value": custom["selected_value"],
                    "price_modifier": custom.get("price_modifier", 0.0)
                }
                await self.db.order_item_customizations.insert_one(customization)
        
        # Generate order token
        token_number = await self._get_next_token_number()
        token, token_prefix = generate_order_token(token_number)
        
        order_token = {
            "id": str(uuid.uuid4()),
            "order_id": order_id,
            "token": token,
            "token_number": token_number,
            "token_prefix": token_prefix,
            "generated_at": datetime.utcnow(),
            "expires_at": datetime.utcnow() + timedelta(days=1)
        }
        await self.db.order_tokens.insert_one(order_token)
        
        # Create payment record
        payment_id = str(uuid.uuid4())
        payment = {
            "id": payment_id,
            "order_id": order_id,
            "method": request.payment_method,
            "status": PaymentStatus.PENDING,
            "amount": cart["total"],
            "upi_id": request.upi_id,
            "card_last_4_digits": request.card_last_4_digits,
            "card_type": request.card_type,
            "initiated_at": datetime.utcnow(),
            "retry_count": 0
        }
        await self.db.payments.insert_one(payment)
        
        # Add timeline entry
        await self._add_timeline_entry(
            order_id,
            OrderStatus.PENDING,
            "Order created",
            user_id,
            user_data.get("name")
        )
        
        # Clear cart
        await self.db.cart_items.delete_many({"cart_id": cart["id"]})
        await self.db.shopping_carts.update_one(
            {"id": cart["id"]},
            {
                "$set": {
                    "subtotal": 0.0,
                    "tax": 0.0,
                    "total": 0.0,
                    "item_count": 0,
                    "updated_at": datetime.utcnow()
                }
            }
        )
        
        # Invalidate caches
        await cache_delete(self.redis, f"cart:{user_id}")
        await cache_delete_pattern(self.redis, f"orders:{user_id}:*")
        
        # Notify queue service to create queue entry
        if auth_token:
            await notify_queue_service(
                order_id=order_id,
                user_id=user_id,
                auth_token=auth_token,
                user_name=user_data.get("name", ""),
                user_phone=user_data.get("phone", ""),
                item_count=cart["item_count"]
            )
        
        return await self.get_order(order_id, user_id)
    
    async def get_order(self, order_id: str, user_id: Optional[str] = None) -> OrderResponse:
        """Get order details"""
        # Try cache first
        cache_key = f"order:{order_id}"
        cached = await cache_get(self.redis, cache_key)
        if cached:
            order = cached
        else:
            query = {"id": order_id}
            if user_id:
                query["user_id"] = user_id
            
            order = await self.db.orders.find_one(query)
            if not order:
                raise ValueError("Order not found")
            
            await cache_set(self.redis, cache_key, order)
        
        # Get order items
        items = []
        async for item in self.db.order_items.find({"order_id": order_id}):
            # Get customizations
            customizations = []
            async for custom in self.db.order_item_customizations.find({"order_item_id": item["id"]}):
                customizations.append(custom)
            item["customizations"] = customizations
            items.append(OrderItemResponse(**item))
        
        # Get token
        token_data = await self.db.order_tokens.find_one({"order_id": order_id})
        token = OrderTokenResponse(**token_data) if token_data else None
        
        # Get payment
        payment_data = await self.db.payments.find_one({"order_id": order_id})
        payment = PaymentResponse(**payment_data) if payment_data else None
        
        # Get timeline
        timeline = []
        async for entry in self.db.order_timeline.find({"order_id": order_id}).sort("timestamp", 1):
            timeline.append(OrderTimelineResponse(**entry))
        
        order["items"] = items
        order["token"] = token
        order["payment"] = payment
        order["timeline"] = timeline
        
        return OrderResponse(**order)
    
    async def get_my_orders(
        self,
        user_id: str,
        status: Optional[OrderStatus] = None,
        page: int = 1,
        page_size: int = 20
    ) -> tuple[List[OrderResponse], int]:
        """Get user's orders with pagination"""
        query = {"user_id": user_id}
        if status:
            query["status"] = status
        
        # Get total count
        total = await self.db.orders.count_documents(query)
        
        # Get orders
        skip = (page - 1) * page_size
        cursor = self.db.orders.find(query).sort("created_at", -1).skip(skip).limit(page_size)
        
        orders = []
        async for order in cursor:
            # Get full order details
            order_response = await self.get_order(order["id"], user_id)
            orders.append(order_response)
        
        return orders, total
    
    async def update_order_status(
        self,
        order_id: str,
        request: UpdateOrderStatusRequest,
        updated_by: str,
        updated_by_name: str,
        is_staff: bool = False
    ) -> OrderResponse:
        """Update order status"""
        order = await self.db.orders.find_one({"id": order_id})
        if not order:
            raise ValueError("Order not found")
        
        # Validate status transitions
        if not is_staff and request.status not in [OrderStatus.CANCELLED]:
            raise ValueError("Only staff can change order status")
        
        update_data = {
            "status": request.status,
            "updated_at": datetime.utcnow()
        }
        
        # Handle specific status updates
        if request.status == OrderStatus.PREPARING:
            if request.estimated_preparation_time:
                update_data["estimated_preparation_time"] = request.estimated_preparation_time
                update_data["estimated_ready_time"] = calculate_estimated_ready_time(
                    request.estimated_preparation_time
                )
        
        elif request.status == OrderStatus.READY:
            update_data["actual_ready_time"] = datetime.utcnow()
            if order.get("estimated_ready_time"):
                start_time = order.get("updated_at", order["created_at"])
                actual_time = (datetime.utcnow() - start_time).total_seconds() / 60
                update_data["actual_preparation_time"] = int(actual_time)
        
        elif request.status == OrderStatus.COMPLETED:
            update_data["completed_at"] = datetime.utcnow()
        
        await self.db.orders.update_one(
            {"id": order_id},
            {"$set": update_data}
        )
        
        # Add timeline entry
        await self._add_timeline_entry(
            order_id,
            request.status,
            request.message or f"Order status changed to {request.status}",
            updated_by,
            updated_by_name
        )
        
        # Invalidate cache
        await cache_delete(self.redis, f"order:{order_id}")
        
        # Notify queue service
        token_data = await self.db.order_tokens.find_one({"order_id": order_id})
        if token_data:
            await notify_queue_service(
                order_id,
                order["order_number"],
                token_data["token"],
                request.status
            )
        
        return await self.get_order(order_id)
    
    async def cancel_order(
        self,
        order_id: str,
        user_id: str,
        request: CancelOrderRequest
    ) -> OrderResponse:
        """Cancel order"""
        order = await self.db.orders.find_one({"id": order_id, "user_id": user_id})
        if not order:
            raise ValueError("Order not found")
        
        # Check if order can be cancelled
        if order["status"] in [OrderStatus.COMPLETED, OrderStatus.CANCELLED, OrderStatus.REFUNDED]:
            raise ValueError("Order cannot be cancelled")
        
        await self.db.orders.update_one(
            {"id": order_id},
            {
                "$set": {
                    "status": OrderStatus.CANCELLED,
                    "cancelled_at": datetime.utcnow(),
                    "cancellation_reason": request.cancellation_reason,
                    "cancelled_by": user_id,
                    "updated_at": datetime.utcnow()
                }
            }
        )
        
        # Update payment status
        await self.db.payments.update_one(
            {"order_id": order_id},
            {"$set": {"status": PaymentStatus.CANCELLED}}
        )
        
        # Add timeline entry
        await self._add_timeline_entry(
            order_id,
            OrderStatus.CANCELLED,
            f"Order cancelled: {request.cancellation_reason}",
            user_id,
            order.get("user_name")
        )
        
        # Invalidate cache
        await cache_delete(self.redis, f"order:{order_id}")
        
        return await self.get_order(order_id, user_id)
    
    async def submit_feedback(
        self,
        order_id: str,
        user_id: str,
        request: OrderFeedbackRequest
    ) -> OrderResponse:
        """Submit order feedback"""
        order = await self.db.orders.find_one({"id": order_id, "user_id": user_id})
        if not order:
            raise ValueError("Order not found")
        
        if order["status"] != OrderStatus.COMPLETED:
            raise ValueError("Can only rate completed orders")
        
        await self.db.orders.update_one(
            {"id": order_id},
            {
                "$set": {
                    "rating": request.rating,
                    "feedback": request.feedback,
                    "feedback_submitted_at": datetime.utcnow(),
                    "updated_at": datetime.utcnow()
                }
            }
        )
        
        # Submit item feedback
        if request.item_feedback:
            for item_feedback in request.item_feedback:
                feedback_entry = {
                    "id": str(uuid.uuid4()),
                    "order_id": order_id,
                    "order_item_id": item_feedback["order_item_id"],
                    "rating": item_feedback["rating"],
                    "comment": item_feedback.get("comment"),
                    "created_at": datetime.utcnow()
                }
                await self.db.order_item_feedback.insert_one(feedback_entry)
        
        # Invalidate cache
        await cache_delete(self.redis, f"order:{order_id}")
        
        return await self.get_order(order_id, user_id)
    
    async def _add_timeline_entry(
        self,
        order_id: str,
        status: OrderStatus,
        message: str,
        updated_by: Optional[str] = None,
        updated_by_name: Optional[str] = None
    ):
        """Add entry to order timeline"""
        entry = {
            "id": str(uuid.uuid4()),
            "order_id": order_id,
            "status": status,
            "message": message,
            "updated_by": updated_by,
            "updated_by_name": updated_by_name,
            "timestamp": datetime.utcnow()
        }
        await self.db.order_timeline.insert_one(entry)
    
    async def _get_next_token_number(self) -> int:
        """Get next token number for the day"""
        today = datetime.utcnow().date()
        
        # Count tokens generated today
        start_of_day = datetime.combine(today, datetime.min.time())
        count = await self.db.order_tokens.count_documents({
            "generated_at": {"$gte": start_of_day}
        })
        
        return count + 1
