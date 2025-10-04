"""
Cart Service - Business logic for shopping cart operations
"""
from typing import Optional, List, Dict, Any
from datetime import datetime
from motor.motor_asyncio import AsyncIOMotorDatabase
from redis.asyncio import Redis
import uuid

from schemas import (
    AddToCartRequest, UpdateCartItemRequest,
    CartResponse, CartItemResponse, CustomizationSchema
)
from utils import (
    calculate_tax, calculate_total, fetch_menu_item_details,
    cache_get, cache_set, cache_delete
)
from config import settings
import logging

logger = logging.getLogger(__name__)


class CartService:
    def __init__(self, db: AsyncIOMotorDatabase, redis: Redis):
        self.db = db
        self.redis = redis
        
    async def get_or_create_cart(self, user_id: str) -> Dict[str, Any]:
        """Get or create shopping cart for user"""
        cart = await self.db.shopping_carts.find_one({"user_id": user_id})
        
        if not cart:
            cart = {
                "id": str(uuid.uuid4()),
                "user_id": user_id,
                "subtotal": 0.0,
                "tax": 0.0,
                "tax_percentage": settings.DEFAULT_TAX_PERCENTAGE,
                "total": 0.0,
                "item_count": 0,
                "created_at": datetime.utcnow(),
                "updated_at": datetime.utcnow()
            }
            await self.db.shopping_carts.insert_one(cart)
            
        return cart
    
    async def get_cart(self, user_id: str) -> Optional[CartResponse]:
        """Get cart with all items"""
        # Try cache first
        cache_key = f"cart:{user_id}"
        cached = await cache_get(self.redis, cache_key)
        if cached:
            return CartResponse(**cached)
        
        cart = await self.get_or_create_cart(user_id)
        
        # Get cart items
        items_cursor = self.db.cart_items.find({"cart_id": cart["id"]})
        items = []
        
        async for item in items_cursor:
            # Get customizations
            customizations_cursor = self.db.cart_item_customizations.find(
                {"cart_item_id": item["id"]}
            )
            customizations = [
                CustomizationSchema(**c) async for c in customizations_cursor
            ]
            
            item["customizations"] = customizations
            items.append(CartItemResponse(**item))
        
        cart["items"] = items
        response = CartResponse(**cart)
        
        # Cache the result
        await cache_set(self.redis, cache_key, response.model_dump())
        
        return response
    
    async def add_to_cart(
        self, 
        user_id: str, 
        request: AddToCartRequest
    ) -> CartResponse:
        """Add item to cart"""
        # Get or create cart
        cart = await self.get_or_create_cart(user_id)
        
        # Fetch menu item details
        menu_item = await fetch_menu_item_details(request.menu_item_id)
        if not menu_item:
            raise ValueError("Menu item not found")
        
        # Check if item already exists in cart
        existing_item = await self.db.cart_items.find_one({
            "cart_id": cart["id"],
            "menu_item_id": request.menu_item_id
        })
        
        if existing_item:
            # Update quantity
            new_quantity = existing_item["quantity"] + request.quantity
            if new_quantity > settings.MAX_QUANTITY_PER_ITEM:
                raise ValueError(f"Maximum quantity per item is {settings.MAX_QUANTITY_PER_ITEM}")
            
            total_price = menu_item["price"] * new_quantity
            
            await self.db.cart_items.update_one(
                {"id": existing_item["id"]},
                {
                    "$set": {
                        "quantity": new_quantity,
                        "total_price": total_price,
                        "updated_at": datetime.utcnow()
                    }
                }
            )
            
            cart_item_id = existing_item["id"]
        else:
            # Create new cart item
            cart_item_id = str(uuid.uuid4())
            total_price = menu_item["price"] * request.quantity
            
            cart_item = {
                "id": cart_item_id,
                "cart_id": cart["id"],
                "menu_item_id": request.menu_item_id,
                "menu_item_name": menu_item["name"],
                "menu_item_image": menu_item.get("image"),
                "quantity": request.quantity,
                "unit_price": menu_item["price"],
                "total_price": total_price,
                "special_instructions": request.special_instructions,
                "added_at": datetime.utcnow(),
                "updated_at": datetime.utcnow()
            }
            
            await self.db.cart_items.insert_one(cart_item)
        
        # Add customizations
        if request.customizations:
            await self.db.cart_item_customizations.delete_many(
                {"cart_item_id": cart_item_id}
            )
            
            for custom in request.customizations:
                customization = {
                    "id": str(uuid.uuid4()),
                    "cart_item_id": cart_item_id,
                    **custom.model_dump()
                }
                await self.db.cart_item_customizations.insert_one(customization)
        
        # Recalculate cart totals
        await self.recalculate_cart(cart["id"])
        
        # Invalidate cache
        await cache_delete(self.redis, f"cart:{user_id}")
        
        return await self.get_cart(user_id)
    
    async def update_cart_item(
        self, 
        user_id: str, 
        cart_item_id: str,
        request: UpdateCartItemRequest
    ) -> CartResponse:
        """Update cart item"""
        cart = await self.get_or_create_cart(user_id)
        
        cart_item = await self.db.cart_items.find_one({
            "id": cart_item_id,
            "cart_id": cart["id"]
        })
        
        if not cart_item:
            raise ValueError("Cart item not found")
        
        update_data = {}
        
        if request.quantity is not None:
            if request.quantity > settings.MAX_QUANTITY_PER_ITEM:
                raise ValueError(f"Maximum quantity per item is {settings.MAX_QUANTITY_PER_ITEM}")
            
            total_price = cart_item["unit_price"] * request.quantity
            update_data["quantity"] = request.quantity
            update_data["total_price"] = total_price
        
        if request.special_instructions is not None:
            update_data["special_instructions"] = request.special_instructions
        
        if update_data:
            update_data["updated_at"] = datetime.utcnow()
            await self.db.cart_items.update_one(
                {"id": cart_item_id},
                {"$set": update_data}
            )
        
        # Update customizations
        if request.customizations is not None:
            await self.db.cart_item_customizations.delete_many(
                {"cart_item_id": cart_item_id}
            )
            
            for custom in request.customizations:
                customization = {
                    "id": str(uuid.uuid4()),
                    "cart_item_id": cart_item_id,
                    **custom.model_dump()
                }
                await self.db.cart_item_customizations.insert_one(customization)
        
        # Recalculate cart totals
        await self.recalculate_cart(cart["id"])
        
        # Invalidate cache
        await cache_delete(self.redis, f"cart:{user_id}")
        
        return await self.get_cart(user_id)
    
    async def remove_cart_item(self, user_id: str, cart_item_id: str) -> CartResponse:
        """Remove item from cart"""
        cart = await self.get_or_create_cart(user_id)
        
        # Delete cart item and its customizations
        result = await self.db.cart_items.delete_one({
            "id": cart_item_id,
            "cart_id": cart["id"]
        })
        
        if result.deleted_count == 0:
            raise ValueError("Cart item not found")
        
        # Recalculate cart totals
        await self.recalculate_cart(cart["id"])
        
        # Invalidate cache
        await cache_delete(self.redis, f"cart:{user_id}")
        
        return await self.get_cart(user_id)
    
    async def clear_cart(self, user_id: str) -> bool:
        """Clear all items from cart"""
        cart = await self.get_or_create_cart(user_id)
        
        # Delete all cart items
        await self.db.cart_items.delete_many({"cart_id": cart["id"]})
        
        # Reset cart totals
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
        
        # Invalidate cache
        await cache_delete(self.redis, f"cart:{user_id}")
        
        return True
    
    async def recalculate_cart(self, cart_id: str):
        """Recalculate cart totals"""
        items_cursor = self.db.cart_items.find({"cart_id": cart_id})
        
        subtotal = 0.0
        item_count = 0
        
        async for item in items_cursor:
            subtotal += item["total_price"]
            item_count += item["quantity"]
        
        tax = calculate_tax(subtotal)
        total = calculate_total(subtotal, tax)
        
        await self.db.shopping_carts.update_one(
            {"id": cart_id},
            {
                "$set": {
                    "subtotal": subtotal,
                    "tax": tax,
                    "total": total,
                    "item_count": item_count,
                    "updated_at": datetime.utcnow()
                }
            }
        )
