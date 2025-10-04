# Testing Guide - API Integration

## Prerequisites

1. All services must be running:

   - Frontend (Next.js): `cd frontend && pnpm dev`
   - Auth Service (Bun): `cd services/auth && bun run src/index.ts`
   - Menu Service (Java): `cd services/menu && ./mvnw spring-boot:run`
   - Order Service (Python): `cd services/order && python main.py`
   - Queue Service (Go): `cd services/queue && go run main.go`
   - Notification Service (Node): `cd services/notification && node server.js`
   - Nginx: `docker-compose -f docker-compose.nginx.yml up`

2. Databases must be running:
   - MySQL (Auth, Menu, Queue)
   - MongoDB (Order, Notification)
   - Redis (Caching, Sessions)
   - Kafka (Message Queue)

## Test Plan

### 1. Staff Dashboard Tests

#### Test 1.1: Page Loads Successfully

**Steps:**

1. Navigate to `http://localhost:3000/staff/dashboard`
2. Login with staff credentials

**Expected Results:**

- ✅ Page loads without errors
- ✅ No "Cannot read properties of undefined" error
- ✅ Statistics cards display with numbers (0 or actual counts)
- ✅ Order tabs render correctly

**API Calls to Verify:**

```
GET /api/admin/orders?status=PENDING
GET /api/admin/orders?status=PENDING&status=CONFIRMED&status=PREPARING&status=READY
```

#### Test 1.2: Statistics Display Correctly

**Steps:**

1. Check the statistics cards at the top

**Expected Results:**

- ✅ Pending count shows actual PENDING orders
- ✅ Preparing count shows actual PREPARING orders
- ✅ Ready count shows actual READY orders
- ✅ Total Orders shows combined count

#### Test 1.3: Tab Navigation

**Steps:**

1. Click on different status tabs (Pending, Confirmed, Preparing, Ready)

**Expected Results:**

- ✅ Tab switches correctly
- ✅ Orders filtered by selected status
- ✅ Order cards display with correct information
- ✅ Each order shows: order number, token, user name, items, total

#### Test 1.4: Order Status Updates

**Steps:**

1. Click "Confirm Order" on a PENDING order
2. Verify order moves to CONFIRMED tab
3. Click "Start Preparing" on a CONFIRMED order
4. Click "Mark as Ready" on a PREPARING order
5. Click "Complete Order" on a READY order

**Expected Results:**

- ✅ Each status update succeeds
- ✅ Order disappears from current tab
- ✅ Order appears in next status tab
- ✅ Toast notification shows success message

**API Calls:**

```
PATCH /api/orders/{id}/status { "status": "CONFIRMED" }
PATCH /api/queue/{id}/status { "status": "IN_PROGRESS" }
```

### 2. Admin Dashboard Tests

#### Test 2.1: Statistics Display

**Steps:**

1. Navigate to `http://localhost:3000/admin/dashboard`
2. Check all statistics cards

**Expected Results:**

- ✅ Total Revenue displays (₹0.00 or actual amount)
- ✅ Orders Today displays (0 or actual count)
- ✅ In Queue displays (0 or actual count)
- ✅ Avg Wait Time displays (0m or actual time)

**API Calls:**

```
GET /api/admin/statistics/today
GET /api/queue/stats
```

#### Test 2.2: Real-time Updates

**Steps:**

1. Keep admin dashboard open
2. Place a new order from another tab/browser
3. Complete an order from staff dashboard

**Expected Results:**

- ✅ Statistics update (may require refresh if WebSocket not working)
- ✅ Numbers reflect new orders
- ✅ Revenue updates correctly

### 3. Admin Orders Page Tests

#### Test 3.1: All Orders Display

**Steps:**

1. Navigate to `http://localhost:3000/admin/orders`

**Expected Results:**

- ✅ All orders in system display (not just admin's orders)
- ✅ Orders from different users visible
- ✅ Each order shows complete information
- ✅ Status badges display correctly

**API Calls:**

```
GET /api/admin/orders
```

#### Test 3.2: Order Management

**Steps:**

1. Update order statuses
2. Cancel orders

**Expected Results:**

- ✅ Status updates work
- ✅ Cancellation works
- ✅ Order list refreshes after actions

### 4. User Orders Page Tests

#### Test 4.1: User's Orders Only

**Steps:**

1. Login as regular user
2. Navigate to `http://localhost:3000/orders`

**Expected Results:**

- ✅ Only user's own orders display
- ✅ No orders from other users visible
- ✅ Order history displayed correctly

**API Calls:**

```
GET /api/orders/my-orders
```

### 5. Analytics Page Tests

#### Test 5.1: Metrics Display

**Steps:**

1. Navigate to `http://localhost:3000/admin/analytics`

**Expected Results:**

- ✅ All metric cards display
- ✅ Order Status Breakdown chart shows percentages
- ✅ Queue Status chart displays
- ✅ All numbers are from real data

**API Calls:**

```
GET /api/admin/statistics/today
GET /api/queue/stats
```

## Manual Testing Scenarios

### Scenario 1: Complete Order Flow

1. **User**: Browse menu, add items to cart
2. **User**: Place order (creates order + queue entry)
3. **Staff**: See order in PENDING tab
4. **Staff**: Confirm order → moves to CONFIRMED
5. **Staff**: Start preparing → moves to PREPARING
6. **Queue**: Entry status updates to IN_PROGRESS
7. **Staff**: Mark as ready → moves to READY
8. **Queue**: Entry status updates to READY
9. **User**: Pick up order
10. **Staff**: Complete order → moves to COMPLETED
11. **Queue**: Entry status updates to COMPLETED
12. **Statistics**: All counts and revenue update

### Scenario 2: Multiple Concurrent Orders

1. Place 5 orders from different users
2. Check staff dashboard shows all 5 in PENDING
3. Confirm 3 orders
4. Check PENDING shows 2, CONFIRMED shows 3
5. Start preparing 2 orders
6. Check statistics update correctly
7. Complete all orders one by one
8. Verify final statistics are accurate

### Scenario 3: Order Cancellation

1. Place order
2. User requests cancellation
3. Staff cancels order
4. Verify order status = CANCELLED
5. Verify order removed from active queue
6. Verify statistics don't count cancelled order in revenue

## Debugging Commands

### Check Order Service Logs

```bash
cd services/order
tail -f logs/order-service.log
```

### Check Queue Service Logs

```bash
cd services/queue
# Logs output to console
```

### Check Database Data

#### MongoDB (Orders)

```bash
docker exec -it mongodb-dev mongosh
use order_db
db.orders.find().pretty()
db.order_statistics.find().pretty()
```

#### MySQL (Queue)

```bash
docker exec -it mysql-dev mysql -u root -proot queue_db
SELECT * FROM queue_entries ORDER BY created_at DESC LIMIT 10;
SELECT * FROM queue_statistics WHERE date = CURDATE();
```

### Test API Endpoints Directly

#### Get Orders (Staff)

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/admin/orders
```

#### Get Statistics

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/admin/statistics/today
```

#### Get Queue Stats

```bash
curl http://localhost:8080/api/queue/stats
```

## Common Issues & Solutions

### Issue 1: "Cannot read properties of undefined (reading 'filter')"

**Solution:** ✅ FIXED - Updated to use `useAllOrders` and access `data.items`

### Issue 2: Staff dashboard shows no orders

**Possible Causes:**

- No orders in the system → Create test orders
- Staff not logged in → Login with staff credentials
- Backend service not running → Check service status
- Database connection issue → Verify DB containers running

**Debug:**

```bash
# Check if orders exist
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/admin/orders
```

### Issue 3: Statistics show 0 for everything

**Possible Causes:**

- No orders created today → Create orders
- Wrong date range → Check system date
- Backend calculation error → Check service logs

**Debug:**

```bash
# Check MongoDB for orders
docker exec -it mongodb-dev mongosh order_db
db.orders.find({ created_at: { $gte: new Date(new Date().setHours(0,0,0,0)) } }).count()

# Check statistics endpoint
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/admin/statistics/today
```

### Issue 4: Order status updates don't work

**Possible Causes:**

- User doesn't have staff role → Check JWT token claims
- Order service not running → Start order service
- Database connection issue → Check logs

**Debug:**

```bash
# Verify staff role in token
# Decode JWT at jwt.io
# Check "roles" array includes "staff" or "admin"

# Try updating directly
curl -X PATCH \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"CONFIRMED"}' \
  http://localhost:8080/api/orders/ORDER_ID/status
```

### Issue 5: Queue statistics not updating

**Possible Causes:**

- Queue service not running → Start queue service
- Statistics calculation job not running → Check service logs
- Database not connected → Verify MySQL container

**Debug:**

```bash
# Check queue service status
curl http://localhost:8080/api/queue/stats

# Check queue entries
curl http://localhost:8080/api/queue
```

## Performance Testing

### Load Test: Multiple Orders

```bash
# Use Apache Bench or k6
ab -n 100 -c 10 -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/admin/orders
```

### Expected Performance

- Order list: < 500ms for 100 orders
- Statistics: < 200ms
- Queue stats: < 100ms
- Order status update: < 300ms

## Success Criteria

✅ All pages load without errors
✅ All API calls return valid data structures
✅ Order flow works end-to-end
✅ Statistics calculate correctly
✅ Staff can manage all orders
✅ Users can only see their own orders
✅ Queue updates sync with order status
✅ WebSocket notifications work (if implemented)

## Next Steps After Testing

1. ✅ Fix any remaining API integration issues
2. ⏭️ Implement WebSocket for real-time updates
3. ⏭️ Add error handling for network failures
4. ⏭️ Implement retry logic for failed requests
5. ⏭️ Add loading states for better UX
6. ⏭️ Implement optimistic updates
7. ⏭️ Add request caching strategies
8. ⏭️ Performance optimization
