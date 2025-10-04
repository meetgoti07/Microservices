# Cart and Checkout System - Implementation Summary

## Overview

Implemented a proper cart and checkout system that syncs with the backend API, fixing the schema mismatch issues.

## Changes Made

### 1. Created Backend Cart API Hooks (`frontend/hooks/swr/useCart.ts`)

**New Hooks:**

- `useCart()` - Fetch user's cart from backend
- `useAddToCart()` - Add item to backend cart
- `useUpdateCartItem()` - Update cart item quantity
- `useRemoveCartItem()` - Remove item from cart
- `useClearCart()` - Clear entire cart

**Key Features:**

- Proper TypeScript types matching backend schema
- Snake_case field names (backend expects `menu_item_id`, not `menuItemId`)
- Automatic cart revalidation after mutations
- Error handling for all operations

### 2. Updated Menu Item Card (`frontend/components/menu/menu-item-card.tsx`)

**Changes:**

- Replaced local cart store (`useCartStore`) with backend API (`useAddToCart`)
- Items now sync to backend immediately when added
- Proper loading states during API calls
- Better error handling with toast notifications

**API Call:**

```typescript
await addToCart({
  menu_item_id: item.id,
  quantity: quantity,
});
```

### 3. Updated Cart Page (`frontend/app/cart/page.tsx`)

**Major Changes:**

- Replaced local cart store with backend cart API
- All cart operations now sync with backend:
  - Update quantity → `PATCH /api/cart/items/{id}`
  - Remove item → `DELETE /api/cart/items/{id}`
  - Clear cart → `DELETE /api/cart`
- Fixed field names to match backend (snake_case)
- Added loading state
- Proper cart totals from backend (includes tax calculation)

**Cart Structure:**

```typescript
{
  items: CartItem[];       // Array of cart items
  subtotal: number;        // Sum of item prices
  tax: number;             // Calculated tax
  tax_percentage: number;  // Tax rate
  total: number;           // Subtotal + tax
  item_count: number;      // Total items
}
```

### 4. Fixed Order Creation (`frontend/hooks/swr/useOrder.ts`)

**Fixed Schema:**

- Changed from camelCase to snake_case
- Removed unnecessary `items` array (backend reads from cart)
- Simplified payload to match backend expectations

**Before:**

```typescript
{
  items: [...],           // ❌ Not needed
  paymentMethod: "CASH"   // ❌ Wrong case
}
```

**After:**

```typescript
{
  payment_method: "CASH",        // ✅ Correct
  special_instructions: undefined // ✅ Optional
}
```

## Backend Flow

### Adding Items to Cart

1. User clicks "Add to Cart" on menu item
2. Frontend → `POST /api/cart/items`
   ```json
   {
     "menu_item_id": "item-123",
     "quantity": 2
   }
   ```
3. Backend:
   - Creates/updates cart in MongoDB
   - Calculates prices and totals
   - Returns updated cart
4. Frontend displays updated cart

### Creating Order

1. User clicks "Proceed to Checkout"
2. Frontend → `POST /api/orders`
   ```json
   {
     "payment_method": "CASH",
     "special_instructions": "Extra spicy"
   }
   ```
3. Backend:
   - Reads items from user's cart in MongoDB
   - Creates order with all cart items
   - Generates order number and token
   - Creates queue entry
   - Clears cart
   - Returns order details
4. Frontend redirects to order page

## Key Backend Endpoints

### Cart Endpoints

- `GET /api/cart` - Get user's cart
- `POST /api/cart/items` - Add item to cart
- `PATCH /api/cart/items/{id}` - Update cart item
- `DELETE /api/cart/items/{id}` - Remove item
- `DELETE /api/cart` - Clear cart

### Order Endpoints

- `POST /api/orders` - Create order from cart
- `GET /api/orders/my-orders` - Get user's orders
- `GET /api/orders/{id}` - Get single order

## Schema Alignment

### Backend Expects (Python/FastAPI)

```python
class CreateOrderRequest(BaseModel):
    payment_method: PaymentMethod  # snake_case
    special_instructions: Optional[str] = None
    upi_id: Optional[str] = None
    card_last_4_digits: Optional[str] = None
```

### Frontend Sends (TypeScript)

```typescript
{
  payment_method: string;  // ✅ Matches
  special_instructions?: string;  // ✅ Matches
}
```

## Error Handling

### Before (Error)

```json
{
  "detail": [
    {
      "type": "missing",
      "loc": ["body", "payment_method"],
      "msg": "Field required"
    }
  ]
}
```

### After (Success)

```json
{
  "id": "order-123",
  "order_number": "ORD-20251003-001",
  "status": "PENDING",
  "token": { "token": "A001" },
  ...
}
```

## Benefits

1. **Data Consistency**: Cart synced across devices/sessions
2. **Real-time Updates**: Cart changes immediately reflected
3. **Server-side Validation**: Backend validates all items
4. **Accurate Pricing**: Tax and totals calculated server-side
5. **Order Reliability**: Orders created from verified cart data
6. **Better UX**: Loading states and error handling

## Testing Checklist

- [x] Add items to cart from menu
- [x] Update cart item quantities
- [x] Remove items from cart
- [x] Clear entire cart
- [x] View cart totals (subtotal, tax, total)
- [x] Select payment method
- [x] Create order from cart
- [x] Order redirects to order page
- [x] Cart clears after order placement
- [x] Handle errors gracefully

## Files Modified

1. `frontend/hooks/swr/useCart.ts` - NEW: Backend cart API hooks
2. `frontend/hooks/swr/useOrder.ts` - Fixed createOrder schema
3. `frontend/components/menu/menu-item-card.tsx` - Use backend cart API
4. `frontend/app/cart/page.tsx` - Complete rewrite with backend integration

## Known Issues Fixed

✅ **Schema Mismatch**: Fixed camelCase vs snake_case
✅ **Missing payment_method**: Now sends correct field name
✅ **Unnecessary items array**: Removed from order creation
✅ **Cart not persisting**: Now stored in backend database
✅ **Tax calculation inconsistency**: Now calculated server-side

## Next Steps

1. Add payment gateway integration
2. Implement order tracking
3. Add order history pagination
4. Enable order modifications (before preparation)
5. Add cart item customizations UI
6. Implement special instructions UI
7. Add cart session timeout handling
8. Implement cart migration for logged-in users

## API Documentation

See `API_ENDPOINTS_REFERENCE.md` for complete API documentation.
