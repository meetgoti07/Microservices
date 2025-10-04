# Queue Integration Fix

## Issues Fixed

### 1. Admin/Staff Role Authorization (403 Forbidden)

**Problem**: Admin endpoints were returning 403 Forbidden even with valid admin JWT token.

**Root Cause**: JWT middleware was checking for `roles` array but the JWT token contains a single `role` string field.

**Files Modified**:

- `services/order/routers/admin.py` - Fixed `require_staff()` function
- `services/order/routers/orders.py` - Fixed `check_staff_role()` function

**Changes**:

```python
# Before
def require_staff(user: Dict[str, Any]) -> Dict[str, Any]:
    roles = user.get("roles", [])
    if "staff" not in roles and "admin" not in roles:
        raise HTTPException(...)
    return user

# After
def require_staff(user: Dict[str, Any]) -> Dict[str, Any]:
    role = user.get("role", "")
    roles = user.get("roles", [])
    # Check both single role string and roles array for compatibility
    if role not in ["staff", "admin"] and "staff" not in roles and "admin" not in roles:
        raise HTTPException(...)
    return user
```

### 2. Orders Page - Snake Case to Camel Case Transformation

**Problem**: Frontend was getting undefined errors when accessing `order.items.slice()` because backend returns `menu_item_name` but frontend expects `menuItemName`.

**Root Cause**: No transformation between backend snake_case and frontend camelCase.

**Files Modified**:

- `frontend/hooks/swr/useOrder.ts`

**Changes**:

- Added `transformKeys()` function to recursively convert snake_case to camelCase
- Updated `useOrders()`, `useAllOrders()`, and `useOrder()` to transform responses
- Added safety check in `frontend/app/orders/page.tsx` for undefined items

**Impact**: All order data now properly displays with correct field names (menuItemName, menuItemId, etc.)

### 3. Queue Service Integration

**Problem**: Queue entries weren't being created when orders were placed. Logs showed `404 POST /api/queue/orders`.

**Root Cause**:

1. Order service was calling wrong endpoint (`/api/queue/orders` instead of `/api/queue`)
2. Order service was sending wrong payload format (missing `user_id` which is required)

**Files Modified**:

- `services/order/utils.py` - Updated `notify_queue_service()` function
- `services/order/services/order_service.py` - Updated function call

**Changes**:

```python
# Before
async def notify_queue_service(order_id: str, order_number: str, token: str, status: str):
    await client.post(
        f"{settings.QUEUE_SERVICE_URL}/api/queue/orders",
        json={
            "order_id": order_id,
            "order_number": order_number,
            "token": token,
            "status": status
        }
    )

# After
async def notify_queue_service(
    order_id: str,
    user_id: str,
    user_name: str = "",
    user_phone: str = "",
    item_count: int = 0
):
    await client.post(
        f"{settings.QUEUE_SERVICE_URL}/api/queue",
        json={
            "order_id": order_id,
            "user_id": user_id,  # Required field
            "user_name": user_name,
            "user_phone": user_phone,
            "item_count": item_count,
            "priority": "NORMAL"
        }
    )
```

### 4. Order Service - Full Order Details in List

**Problem**: `my-orders` endpoint was returning `OrderListResponse` with only `item_count`, no actual items array.

**Root Cause**: `get_my_orders()` was using a lightweight response schema for performance but frontend needed full details.

**Files Modified**:

- `services/order/services/order_service.py` - Changed `get_my_orders()` return type

**Changes**:

```python
# Before
async def get_my_orders(...) -> tuple[List[OrderListResponse], int]:
    # Only returned order summary with item_count

# After
async def get_my_orders(...) -> tuple[List[OrderResponse], int]:
    # Returns full order details including items array
    for order in cursor:
        order_response = await self.get_order(order["id"], user_id)
        orders.append(order_response)
```

## Queue Service Expected Request Format

When creating a queue entry, the queue service expects:

**Endpoint**: `POST /api/queue`

**Required Headers**:

```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request Body**:

```json
{
  "order_id": "string (required)",
  "user_id": "string (required)",
  "user_name": "string (optional)",
  "user_phone": "string (optional)",
  "item_count": 0,
  "priority": "NORMAL|HIGH|LOW"
}
```

**Response**:

```json
{
  "message": "Queue entry created successfully",
  "data": {
    "id": "uuid",
    "order_id": "uuid",
    "user_id": "uuid",
    "token": "A001",
    "position": 1,
    "status": "WAITING",
    "created_at": "2025-10-04T00:00:00Z"
  }
}
```

## Testing Checklist

- [x] Admin can access `/api/admin/orders` without 403
- [x] Admin can access `/api/admin/statistics/today` without 403
- [x] Staff can update order status without 403
- [x] Orders page displays items correctly with proper field names
- [ ] Creating an order creates a queue entry
- [ ] Queue entry appears in queue display
- [ ] Queue statistics update properly
- [ ] Real-time notifications work when queue updates

## Next Steps

1. **Restart Order Service**: Apply Python code changes

   ```bash
   cd services/order
   python main.py
   ```

2. **Test Order → Queue Flow**:

   - Create an order from frontend
   - Check queue service logs for queue entry creation
   - Verify queue entry appears in `/queue` page
   - Confirm user receives token notification

3. **Test Notification System**:
   - Ensure notification service is running
   - Verify WebSocket connections work
   - Test real-time updates when queue advances

## Files Changed Summary

1. `services/order/routers/admin.py` - Role checking
2. `services/order/routers/orders.py` - Role checking
3. `services/order/services/order_service.py` - Queue notification & full order details
4. `services/order/utils.py` - Queue service integration
5. `frontend/hooks/swr/useOrder.ts` - Case transformation
6. `frontend/app/orders/page.tsx` - Safety checks

## Architecture Notes

```
Order Service (Python)
  ↓ (when order created)
  POST /api/queue
  ↓
Queue Service (Go)
  → Creates queue entry
  → Generates token
  → Updates position
  → Publishes Kafka event
  ↓
Notification Service (Node.js)
  → Listens to Kafka
  → Sends WebSocket notification
  → Updates Redis cache
  ↓
Frontend (Next.js)
  → Receives real-time update
  → Shows token & position
  → Updates queue display
```
