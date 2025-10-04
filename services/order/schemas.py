"""
Pydantic schemas for type-safe API requests and responses
"""
from pydantic import BaseModel, Field, validator, ConfigDict
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum
from decimal import Decimal


# Enums
class OrderStatus(str, Enum):
    CART = "CART"
    PENDING = "PENDING"
    CONFIRMED = "CONFIRMED"
    PREPARING = "PREPARING"
    READY = "READY"
    COMPLETED = "COMPLETED"
    CANCELLED = "CANCELLED"
    REFUNDED = "REFUNDED"


class PaymentMethod(str, Enum):
    CASH = "CASH"
    UPI = "UPI"
    CARD = "CARD"
    WALLET = "WALLET"
    CAMPUS_CARD = "CAMPUS_CARD"
    MEAL_COUPON = "MEAL_COUPON"


class PaymentStatus(str, Enum):
    PENDING = "PENDING"
    PROCESSING = "PROCESSING"
    COMPLETED = "COMPLETED"
    FAILED = "FAILED"
    REFUNDED = "REFUNDED"
    CANCELLED = "CANCELLED"


class CardType(str, Enum):
    VISA = "VISA"
    MASTERCARD = "MASTERCARD"
    RUPAY = "RUPAY"
    AMEX = "AMEX"


# Base Schemas
class CustomizationSchema(BaseModel):
    customization_id: str
    customization_name: str
    selected_label: str
    selected_value: str
    price_modifier: float = 0.0


# Cart Schemas
class AddToCartRequest(BaseModel):
    menu_item_id: str
    quantity: int = Field(ge=1, le=50)
    special_instructions: Optional[str] = None
    customizations: Optional[List[CustomizationSchema]] = []


class UpdateCartItemRequest(BaseModel):
    quantity: Optional[int] = Field(None, ge=1, le=50)
    special_instructions: Optional[str] = None
    customizations: Optional[List[CustomizationSchema]] = None


class CartItemResponse(BaseModel):
    id: str
    cart_id: str
    menu_item_id: str
    menu_item_name: str
    menu_item_image: Optional[str] = None
    quantity: int
    unit_price: float
    total_price: float
    special_instructions: Optional[str] = None
    customizations: List[CustomizationSchema] = []
    added_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)


class CartResponse(BaseModel):
    id: str
    user_id: str
    subtotal: float
    tax: float
    tax_percentage: float
    total: float
    item_count: int
    items: List[CartItemResponse] = []
    created_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)


# Order Schemas
class CreateOrderRequest(BaseModel):
    payment_method: PaymentMethod
    special_instructions: Optional[str] = None
    upi_id: Optional[str] = None
    card_last_4_digits: Optional[str] = None
    card_type: Optional[CardType] = None


class OrderItemResponse(BaseModel):
    id: str
    order_id: str
    menu_item_id: str
    menu_item_name: str
    menu_item_image: Optional[str] = None
    quantity: int
    unit_price: float
    total_price: float
    special_instructions: Optional[str] = None
    customizations: List[CustomizationSchema] = []

    model_config = ConfigDict(from_attributes=True)


class OrderTokenResponse(BaseModel):
    id: str
    order_id: str
    token: str
    token_number: int
    token_prefix: Optional[str] = None
    generated_at: datetime
    expires_at: Optional[datetime] = None

    model_config = ConfigDict(from_attributes=True)


class OrderTimelineResponse(BaseModel):
    id: str
    order_id: str
    status: OrderStatus
    message: Optional[str] = None
    updated_by: Optional[str] = None
    updated_by_name: Optional[str] = None
    timestamp: datetime
    metadata: Optional[Dict[str, Any]] = None

    model_config = ConfigDict(from_attributes=True)


class PaymentResponse(BaseModel):
    id: str
    order_id: str
    method: PaymentMethod
    status: PaymentStatus
    amount: float
    transaction_id: Optional[str] = None
    transaction_reference: Optional[str] = None
    payment_gateway: Optional[str] = None
    upi_id: Optional[str] = None
    card_last_4_digits: Optional[str] = None
    card_type: Optional[CardType] = None
    initiated_at: datetime
    completed_at: Optional[datetime] = None
    failure_reason: Optional[str] = None

    model_config = ConfigDict(from_attributes=True)


class OrderResponse(BaseModel):
    id: str
    order_number: str
    user_id: str
    user_name: Optional[str] = None
    user_email: Optional[str] = None
    user_phone: Optional[str] = None
    status: OrderStatus
    subtotal: float
    tax: float
    total: float
    special_instructions: Optional[str] = None
    estimated_preparation_time: int = 0
    actual_preparation_time: Optional[int] = None
    estimated_ready_time: Optional[datetime] = None
    actual_ready_time: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    cancelled_at: Optional[datetime] = None
    cancellation_reason: Optional[str] = None
    rating: Optional[int] = None
    feedback: Optional[str] = None
    items: List[OrderItemResponse] = []
    token: Optional[OrderTokenResponse] = None
    payment: Optional[PaymentResponse] = None
    timeline: List[OrderTimelineResponse] = []
    created_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)


class OrderListResponse(BaseModel):
    id: str
    order_number: str
    status: OrderStatus
    total: float
    item_count: int
    estimated_ready_time: Optional[datetime] = None
    token: Optional[str] = None
    created_at: datetime

    model_config = ConfigDict(from_attributes=True)


class UpdateOrderStatusRequest(BaseModel):
    status: OrderStatus
    message: Optional[str] = None
    estimated_preparation_time: Optional[int] = None


class CancelOrderRequest(BaseModel):
    cancellation_reason: str


class OrderFeedbackRequest(BaseModel):
    rating: int = Field(ge=1, le=5)
    feedback: Optional[str] = None
    item_feedback: Optional[List[Dict[str, Any]]] = None


class OrderItemFeedbackRequest(BaseModel):
    order_item_id: str
    rating: int = Field(ge=1, le=5)
    comment: Optional[str] = None


# Statistics Schemas
class OrderStatisticsResponse(BaseModel):
    date: str
    total_orders: int
    pending_orders: int
    confirmed_orders: int
    preparing_orders: int
    ready_orders: int
    completed_orders: int
    cancelled_orders: int
    refunded_orders: int
    total_revenue: float
    total_tax: float
    avg_order_value: float
    avg_preparation_time: int

    model_config = ConfigDict(from_attributes=True)


# Pagination
class PaginatedResponse(BaseModel):
    items: List[Any]
    total: int
    page: int
    page_size: int
    total_pages: int
    has_next: bool
    has_prev: bool


# Response wrappers
class SuccessResponse(BaseModel):
    success: bool = True
    message: str
    data: Optional[Any] = None


class ErrorResponse(BaseModel):
    success: bool = False
    error: str
    details: Optional[Any] = None
