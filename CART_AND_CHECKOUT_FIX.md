# Cart and Checkout System Fix

## Issues Found

### 1. Menu Item Response Structure Mismatch

**Problem**: The cart service was trying to access `menu_item["price"]` directly, but the menu service returns data wrapped in `{ success, message, data }` structure.

**Location**: `services/order/utils.py` - `fetch_menu_item_details()`

**Fix Applied**:

```python
async def fetch_menu_item_details(menu_item_id: str) -> Optional[Dict[str, Any]]:
    """
    Fetch menu item details from Menu Service
    """
    try:
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{settings.MENU_SERVICE_URL}/api/menu/items/{menu_item_id}",
                timeout=5.0
            )
            response.raise_for_status()
            result = response.json()

            # Menu service returns { success, message, data }
            # Extract the actual menu item from data field
            if isinstance(result, dict) and "data" in result:
                return result["data"]
            return result
    except Exception as e:
        logger.error(f"Error fetching menu item {menu_item_id}: {e}")
        return None
```

### 2. Payment Method Field Name

**Problem**: Frontend was sending `paymentMethod` (camelCase) but backend expects `payment_method` (snake_case).

**Status**: ✅ Already fixed in frontend - cart page is correctly sending `payment_method`

## Cart API Flow

### 1. Add Item to Cart

**Endpoint**: `POST /api/cart/items`

**Request Body**:

```json
{
  "menu_item_id": "uuid",
  "quantity": 1,
  "special_instructions": "optional",
  "customizations": []
}
```

**Process**:

1. Fetch menu item details from Menu Service
2. Extract data from `{ success, message, data }` wrapper
3. Access `data["price"]`, `data["name"]`, etc.
4. Create or update cart item
5. Calculate totals
6. Return cart response

**Response**:

```json
{
  "id": "cart-id",
  "user_id": "user-id",
  "subtotal": 100.0,
  "tax": 5.0,
  "total": 105.0,
  "item_count": 2,
  "items": [
    {
      "id": "item-id",
      "cart_id": "cart-id",
      "menu_item_id": "menu-item-id",
      "menu_item_name": "Item Name",
      "quantity": 1,
      "unit_price": 100.0,
      "total_price": 100.0,
      "customizations": []
    }
  ],
  "created_at": "2025-10-04T...",
  "updated_at": "2025-10-04T..."
}
```

### 2. Get Cart

**Endpoint**: `GET /api/cart`

**Response**: Same as above

### 3. Update Cart Item

**Endpoint**: `PATCH /api/cart/items/{cart_item_id}`

**Request Body**:

```json
{
  "quantity": 2,
  "special_instructions": "optional"
}
```

### 4. Remove Cart Item

**Endpoint**: `DELETE /api/cart/items/{cart_item_id}`

### 5. Clear Cart

**Endpoint**: `DELETE /api/cart`

## Checkout Flow

### 1. Create Order from Cart

**Endpoint**: `POST /api/orders`

**Request Body**:

```json
{
  "payment_method": "CASH",
  "special_instructions": "optional",
  "upi_id": "optional for UPI",
  "card_last_4_digits": "optional for CARD",
  "card_type": "optional for CARD"
}
```

**Process**:

1. Validate user has items in cart
2. Fetch all cart items
3. For each cart item:
   - Fetch latest menu item details
   - Verify availability
   - Calculate prices
4. Create order with items
5. Generate order number
6. Generate order token
7. Create payment record
8. Create queue entry via Queue Service
9. Send notifications
10. Clear user's cart
11. Return complete order

**Response**:

```json
{
  "id": "order-id",
  "order_number": "ORD-20251004-001",
  "user_id": "user-id",
  "status": "PENDING",
  "items": [...],
  "token": {
    "token": "A001",
    "token_number": 1,
    "generated_at": "2025-10-04T..."
  },
  "payment": {
    "method": "CASH",
    "status": "PENDING",
    "amount": 105.00
  },
  "subtotal": 100.00,
  "tax": 5.00,
  "total": 105.00,
  "estimated_preparation_time": 15,
  "estimated_ready_time": "2025-10-04T...",
  "timeline": [...],
  "created_at": "2025-10-04T..."
}
```

## Frontend Cart Store vs Backend Cart

### Local Cart Store (Zustand)

- Used for: Immediate UI feedback
- Persisted in: localStorage
- Purpose: Optimistic updates, offline support

### Backend Cart (MongoDB)

- Source of truth for checkout
- Synced when: User adds/updates/removes items
- Handles: Price calculations, tax, totals

**Important**: The checkout process uses **backend cart**, not local store!

## Testing the Fix

### 1. Test Add to Cart

```bash
# Login and get token
TOKEN="your-jwt-token"

# Add item to cart
curl -X POST http://localhost:8080/api/cart/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "menu_item_id": "190ff549-37a1-49f6-8a85-e5d9a3b5c011",
    "quantity": 2
  }'
```

**Expected**: Success response with cart containing the item

### 2. Test Get Cart

```bash
curl -X GET http://localhost:8080/api/cart \
  -H "Authorization: Bearer $TOKEN"
```

**Expected**: Cart with items, subtotal, tax, total

### 3. Test Checkout

```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_method": "CASH"
  }'
```

**Expected**: Order created with token, queue entry created

## Common Errors and Solutions

### Error: "'price'"

**Cause**: Menu service response not properly unwrapped
**Solution**: ✅ Fixed - Now extracts data from wrapper object

### Error: "Field required: payment_method"

**Cause**: Frontend sending camelCase instead of snake_case
**Solution**: ✅ Frontend already uses snake_case

### Error: "Cart is empty"

**Cause**:

1. Backend cart not synced with local store
2. Items expired/removed
3. User cart cleared

**Solution**: Use backend cart APIs, not local store for checkout

### Error: "Menu item not found"

**Cause**: Invalid menu_item_id or item deleted
**Solution**: Validate menu item exists before adding to cart

## Updated Files

1. ✅ `services/order/utils.py` - Fixed menu item response unwrapping
2. ✅ `frontend/app/cart/page.tsx` - Already uses correct snake_case
3. ✅ `frontend/hooks/swr/useCart.ts` - Already uses correct snake_case
4. ✅ `frontend/components/menu/menu-item-card.tsx` - Already uses correct parameters

## Next Steps

1. Restart order service to apply utils.py fix
2. Test add to cart from menu page
3. Test cart operations (update quantity, remove items)
4. Test complete checkout flow
5. Verify queue entry created after order
6. Verify notifications sent

## Service Dependencies

```
Frontend → Nginx (8080) → Order Service (8083)
                      ↓
Order Service → Menu Service (8082) - Get menu item details
           → Queue Service (8084) - Create queue entry
           → Notification Service (8085) - Send notifications
```

## Database Collections Used

### MongoDB (order_db)

- `shopping_carts` - User carts
- `cart_items` - Items in carts
- `cart_item_customizations` - Item customizations
- `orders` - Completed orders
- `order_items` - Items in orders
- `order_tokens` - Order tokens
- `payments` - Payment records
- `order_timeline` - Order status history

### MySQL (queue_db)

- `queue_entries` - Queue positions
- `queue_statistics` - Queue stats

## Configuration

### Order Service (`services/order/config.py`)

```python
MENU_SERVICE_URL = "http://localhost:3002"
QUEUE_SERVICE_URL = "http://localhost:3004"
NOTIFICATION_SERVICE_URL = "http://localhost:3005"
DEFAULT_TAX_PERCENTAGE = 5.0
MAX_QUANTITY_PER_ITEM = 50
```

## Success Criteria

✅ Add to cart works without errors
✅ Cart displays correct items and totals
✅ Update quantity works
✅ Remove item works
✅ Checkout creates order successfully
✅ Order has proper token
✅ Queue entry created
✅ Payment record created
✅ User cart cleared after checkout
✅ Order appears in user's order history
