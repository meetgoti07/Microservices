# API Endpoints Reference Guide

## Base URLs

- **Frontend**: http://localhost:3000
- **API Gateway (Nginx)**: http://localhost:8080
- **Auth Service**: http://localhost:3001 (internal)
- **Menu Service**: http://localhost:8082 (internal)
- **Order Service**: http://localhost:8083 (internal)
- **Queue Service**: http://localhost:8084 (internal)
- **Notification Service**: http://localhost:8085 (internal)

All frontend requests go through Nginx at `http://localhost:8080`

## Authentication Endpoints

### Auth Service - `/api/auth`

#### Public Routes

- `POST /api/auth/sign-in` - User login
- `POST /api/auth/sign-up` - User registration
- `GET /api/auth/session` - Get current session

#### Protected Routes

- `POST /api/auth/sign-out` - User logout
- `GET /api/auth/user` - Get current user info

## Menu Endpoints

### Menu Service - `/api/menu`

#### Public Routes

- `GET /api/menu/items` - Get all menu items (with filters)
- `GET /api/menu/items/{id}` - Get single menu item
- `GET /api/menu/categories` - Get all categories

#### Admin Routes (require admin role)

- `POST /api/menu/items` - Create menu item
- `PUT /api/menu/items/{id}` - Update menu item
- `DELETE /api/menu/items/{id}` - Delete menu item
- `POST /api/menu/categories` - Create category
- `PUT /api/menu/categories/{id}` - Update category
- `DELETE /api/menu/categories/{id}` - Delete category

## Order Endpoints

### Order Service - `/api/orders`

#### User Routes (require authentication)

- `POST /api/orders` - Create order from cart
  - Request: `{ payment_method, special_instructions?, ... }`
  - Response: Full order object with token
- `GET /api/orders/my-orders` - Get user's orders (paginated)
  - Query params: `status`, `page`, `page_size`
  - Response: `{ items: Order[], total, page, ... }`
- `GET /api/orders/{id}` - Get single order
  - Returns: Full order with items, payment, timeline
- `POST /api/orders/{id}/cancel` - Cancel order
  - Request: `{ cancellation_reason }`
- `POST /api/orders/{id}/feedback` - Submit order feedback
  - Request: `{ rating, feedback?, item_feedback? }`

#### Staff Routes (require staff/admin role)

- `PATCH /api/orders/{id}/status` - Update order status
  - Request: `{ status, message?, estimated_preparation_time? }`
  - Also updates queue status automatically

### Admin Routes - `/api/admin`

#### Staff/Admin Routes

- `GET /api/admin/orders` - Get all orders (paginated, with filters)
  - Query params: `status`, `search`, `page`, `page_size`
  - Response: `{ items: Order[], total, page, ... }`
- `GET /api/admin/orders/active` - Get all active orders
  - Returns: Array of orders with status PENDING, CONFIRMED, PREPARING, or READY
- `GET /api/admin/statistics/today` - Get today's statistics
  - Response:
    ```json
    {
      "date": "2025-10-03",
      "total_orders": 45,
      "pending_orders": 5,
      "confirmed_orders": 3,
      "preparing_orders": 7,
      "ready_orders": 2,
      "completed_orders": 28,
      "cancelled_orders": 0,
      "refunded_orders": 0,
      "total_revenue": 12450.5,
      "total_tax": 1245.05,
      "avg_order_value": 276.68,
      "avg_preparation_time": 15
    }
    ```

## Queue Endpoints

### Queue Service - `/api/queue`

#### Public Routes

- `GET /api/queue` - Get all active queue entries
- `GET /api/queue/current` - Get current queue display (for screens)
- `GET /api/queue/stats` - Get queue statistics

  - Optional query: `date` (YYYY-MM-DD format)
  - Response:
    ```json
    {
      "total_in_queue": 10,
      "waiting_count": 4,
      "in_progress_count": 5,
      "ready_count": 1,
      "completed_today": 35,
      "cancelled_today": 2,
      "avg_wait_time": 12,
      "avg_preparation_time": 15,
      "longest_wait_time": 25,
      "shortest_wait_time": 5,
      "current_load": 65.5,
      "peak_load": 85.2
    }
    ```

- `GET /api/queue/position/{token}` - Get queue position by token
- `GET /api/queue/token/{token}` - Get queue entry by token

#### User Routes (require authentication)

- `POST /api/queue` - Create queue entry
- `GET /api/queue/order/{orderId}` - Get queue entry by order ID
- `GET /api/queue/user/me` - Get user's queue entries

#### Staff Routes (require staff/admin role)

- `PATCH /api/queue/{id}/status` - Update queue status
  - Request: `{ status, estimated_ready_time?, ... }`
  - Statuses: WAITING, IN_PROGRESS, READY, COMPLETED, CANCELLED
- `PUT /api/queue/{id}/priority` - Update queue priority
- `POST /api/queue/{id}/assign` - Assign staff to queue entry
- `POST /api/queue/advance` - Advance entire queue
- `GET /api/queue/{id}/logs` - Get staff action logs
- `GET /api/queue/config` - Get queue configuration
- `POST /api/queue/recalculate` - Recalculate all positions

#### Admin Routes

- `PUT /api/queue/config` - Update queue configuration

## Cart Endpoints

### Order Service - `/api/cart`

#### User Routes (require authentication)

- `GET /api/cart` - Get user's cart
  - Response: Full cart with items, subtotal, tax, total
- `POST /api/cart/items` - Add item to cart
  - Request: `{ menu_item_id, quantity, special_instructions?, customizations? }`
- `PATCH /api/cart/items/{id}` - Update cart item
  - Request: `{ quantity?, special_instructions?, customizations? }`
- `DELETE /api/cart/items/{id}` - Remove item from cart
- `DELETE /api/cart` - Clear entire cart

## Notification Endpoints

### Notification Service - `/api`

#### User Routes (require authentication)

- `GET /api/notifications` - Get user's notifications
- `GET /api/notifications/unread` - Get unread notifications
- `PATCH /api/notifications/{id}/read` - Mark notification as read
- `DELETE /api/notifications/{id}` - Delete notification

#### WebSocket

- `ws://localhost:8080/ws` - WebSocket connection for real-time updates
  - Requires authentication token
  - Receives order updates, queue position changes, etc.

## Common Response Structures

### Paginated Response

```json
{
  "items": [...],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_pages": 5,
  "has_next": true,
  "has_prev": false
}
```

### Error Response

```json
{
  "error": "Error message",
  "details": "Additional details",
  "status": 400
}
```

### Success Response

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... }
}
```

## Order Statuses

- `CART` - Items in cart, not yet ordered
- `PENDING` - Order placed, awaiting confirmation
- `CONFIRMED` - Order confirmed by staff
- `PREPARING` - Order being prepared
- `READY` - Order ready for pickup
- `COMPLETED` - Order picked up
- `CANCELLED` - Order cancelled
- `REFUNDED` - Order refunded

## Queue Statuses

- `WAITING` - In queue, waiting
- `IN_PROGRESS` - Being prepared
- `READY` - Ready for pickup
- `COMPLETED` - Picked up
- `CANCELLED` - Cancelled
- `NO_SHOW` - Customer didn't show up
- `EXPIRED` - Token expired

## Frontend Hook Usage

### Orders

```typescript
// Get user's orders
const { data } = useOrders({ status: ["PENDING"] });
// Access: data.items

// Get all orders (staff/admin)
const { data } = useAllOrders({ status: ["PREPARING"] });
// Access: data.items

// Get single order
const { data } = useOrder(orderId);

// Create order
const { trigger } = useCreateOrder();
await trigger({ payment_method: "CASH" });

// Update order status
const { trigger } = useUpdateOrderStatus(orderId);
await trigger({ status: "CONFIRMED" });

// Get statistics
const { data } = useOrderStatistics();
// Access: data.totalRevenue, data.totalOrders, etc.
```

### Queue

```typescript
// Get queue statistics
const { data } = useQueueStatistics();
// Access: data.totalInQueue, data.averageWaitTime, etc.

// Get user's queue position
const { data } = useMyQueuePosition();

// Update queue status (staff)
const { trigger } = useUpdateQueueStatus(queueId);
await trigger({ status: "IN_PROGRESS" });
```

### Menu

```typescript
// Get menu items
const { data } = useMenuItems({ category: "meals" });

// Get categories
const { data } = useCategories();
```

## Authentication Headers

All protected endpoints require JWT token in Authorization header:

```
Authorization: Bearer <token>
```

The token is automatically included by the `apiClient` utility from localStorage.

## Role-Based Access

- **Public**: No authentication required
- **User**: Any authenticated user
- **Staff**: User with 'staff' or 'admin' role
- **Admin**: User with 'admin' role

Roles are checked via JWT middleware in each service.
