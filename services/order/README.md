# Order Service

FastAPI-based microservice for managing orders, shopping carts, and payments.

## Features

- **Shopping Cart Management**

  - Add/update/remove items from cart
  - Calculate totals with tax
  - Item customizations support
  - Redis caching for fast access

- **Order Management**

  - Create orders from cart
  - Order status tracking
  - Order timeline/history
  - Token generation for pickup
  - Cancel orders
  - Submit feedback and ratings

- **Payment Processing**

  - Multiple payment methods (Cash, UPI, Card, Wallet, etc.)
  - Payment status tracking
  - Transaction recording

- **Admin Features**
  - View all orders
  - Update order status
  - View active orders
  - Daily statistics
  - Payment reports

## Tech Stack

- **FastAPI** - Web framework
- **MongoDB** - Primary database
- **Redis** - Caching layer
- **Motor** - Async MongoDB driver
- **Pydantic** - Data validation

## Installation

1. Install dependencies:

```bash
pip install -r requirements.txt
```

2. Set up environment variables:

```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Ensure MongoDB and Redis are running

## Running the Service

```bash
python main.py
```

Or with uvicorn:

```bash
uvicorn main:app --host 0.0.0.0 --port 3003 --reload
```

## API Endpoints

### Cart Endpoints

- `GET /api/cart` - Get user's cart
- `POST /api/cart/items` - Add item to cart
- `PATCH /api/cart/items/{id}` - Update cart item
- `DELETE /api/cart/items/{id}` - Remove cart item
- `DELETE /api/cart` - Clear cart

### Order Endpoints

- `POST /api/orders` - Create order from cart
- `GET /api/orders/my-orders` - Get user's orders (paginated)
- `GET /api/orders/{id}` - Get order details
- `GET /api/orders/{id}/status` - Get order status (public)
- `PATCH /api/orders/{id}/status` - Update order status (staff only)
- `POST /api/orders/{id}/cancel` - Cancel order
- `POST /api/orders/{id}/feedback` - Submit feedback

### Admin Endpoints (Staff/Admin only)

- `GET /api/admin/orders` - Get all orders (paginated, searchable)
- `GET /api/admin/orders/active` - Get active orders
- `GET /api/admin/statistics/today` - Get today's statistics
- `GET /api/admin/statistics/range` - Get statistics for date range
- `GET /api/admin/payments` - Get all payments (admin only)

## Authentication

All endpoints (except public status check) require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Data Models

### Order Status Flow

1. `PENDING` - Order created, payment pending
2. `CONFIRMED` - Payment confirmed
3. `PREPARING` - Order being prepared
4. `READY` - Order ready for pickup
5. `COMPLETED` - Order completed
6. `CANCELLED` - Order cancelled
7. `REFUNDED` - Order refunded

### Payment Methods

- `CASH` - Cash payment
- `UPI` - UPI payment
- `CARD` - Card payment
- `WALLET` - Digital wallet
- `CAMPUS_CARD` - Campus card
- `MEAL_COUPON` - Meal coupon

## Caching Strategy

- Cart data: 5 minutes TTL
- Order data: 5 minutes TTL
- Automatic cache invalidation on updates

## Database Schema

See `migrations/001_create_order_tables.sql` for the complete database schema including:

- Shopping carts and cart items
- Orders and order items
- Order tokens
- Payments
- Order timeline
- Statistics
- Feedback

## Error Handling

The service returns appropriate HTTP status codes:

- `200` - Success
- `201` - Created
- `400` - Bad request (validation error)
- `401` - Unauthorized
- `403` - Forbidden (insufficient permissions)
- `404` - Not found
- `500` - Internal server error

## Development

The service uses:

- Type hints for better IDE support
- Pydantic models for validation
- Modular architecture (routers, services, models)
- Async/await for non-blocking I/O
- Dependency injection

## Integration

This service integrates with:

- **Auth Service** - User authentication and details
- **Menu Service** - Menu item details and pricing
- **Queue Service** - Order queue notifications
