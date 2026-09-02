```markdown
# Bookstore REST API

A lightweight, high-performance RESTful API for managing a bookstore catalog and customer orders built with **Go 1.22+**. Designed following **Clean Architecture (3-Tier Layered Architecture)** principles, featuring thread-safe file persistence, custom HTTP logging, panic recovery, and role-based access control.

---

## Architecture Overview

The application strictly separates concerns into decoupled layers using Go interfaces and Dependency Injection (DI):

```text
[ Client / cURL ]
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│ Global Middleware Chain (internal/middleware)               │
│  • LoggingMiddleware: Captures latency & response details   │
│  • RecoveryMiddleware: Prevents crashes on runtime panics   │
│  • CORSMiddleware: Manages cross-origin headers             │
│  • AuthMiddleware: Validates Bearer JWTs via ValidateToken()│
│  • RequireRole: Enforces RBAC permissions ("ADMIN"/"CUSTOMER")│
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Handler Layer (internal/handler)                            │
│  • HTTP request/response parsing & validation               │
│  • Maps typed apperrors (ErrNotFound, ErrConflict, etc.)   │
│    to HTTP status codes (400, 401, 403, 404, 409, 500)      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Service Layer (internal/service)                            │
│  • Core business logic and validation                       │
│  • Optimistic concurrency control (Version checks)          │
│  • Order price calculation & resource ownership checks      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Repository Layer (internal/repository)                      │
│  • Pluggable storage via interface abstraction              │
│    (BookRepository, UserRepository, OrderRepository)        │
│  • JSON Implementation: Thread-safe file I/O with           │
│    sync.RWMutex (json_*_repository.go)                      │
│  • MongoDB Implementation: Native driver with connection    │
│    pooling & indexes (mongo_*_repository.go)                │
│  • Atomic CAS operations (Version field enforcement)        │
│  • Multi-phase stock validation & deduction                 │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Storage Backends                                            │
│  • MongoDB: Production persistence with indexes & pooling   │
│  • JSON Files (data/): Development/testing alternative      │
└──────────────────────────┴──────────────────────────────────┘

```

---

## Features

* **Clean Architecture:** Strict layer isolation (Handler $\rightarrow$ Service $\rightarrow$ Repository) using Go interfaces and DTO boundaries.
* **Pluggable Storage Backends:** Dual storage support via repository interfaces: production-grade MongoDB with connection pooling (10-100 connections) and indexes, or lightweight JSON files with `sync.RWMutex` for development/testing.
* **Multi-Phase Stock Deduction:** Order processing executes in two distinct phases (Phase 1: In-memory stock validation; Phase 2: Disk mutations) to prevent partial stock loss on validation errors.
* **Global Middleware Pipeline:** Standardized request wrapping using `middleware.Chain` incorporating logging, panic recovery, and CORS headers.
* **Strict Startup Safeguards:** Fail-fast server initialization refusing execution without a configured `JWT_SECRET` environment variable.
* **RESTful CRUD Operations:** Full resource management for users, catalog books, and customer orders.
* **JWT Authentication & RBAC:** Centralized token verification delegating to `auth.ValidateToken()` with role enforcement (`ADMIN` vs. `CUSTOMER`).
* **Optimistic Concurrency Control:** Version-based locking (`version` field) to prevent lost update collisions on concurrent writes.

---

## Project Structure

```text
bookstore-api/
├── cmd/
│   └── api/
│       └── main.go                 # App entrypoint, env verification, DI, & global server startup
├── internal/
│   ├── apperrors/
│   │   └── errors.go               # Custom typed sentinel errors (%w wrapping support)
│   ├── auth/
│   │   └── jwt.go                  # JWT token creation and ValidateToken implementation
│   ├── dto/
│   │   ├── auth_dto.go             # DTOs for user registration and authentication
│   │   ├── book_dto.go             # DTOs for book catalog creation, update, and response
│   │   └── order_dto.go            # DTOs for order placement and fulfillment views
│   ├── handler/
│   │   ├── auth_handler.go         # HTTP handlers for auth routes
│   │   ├── book_handler.go         # HTTP handlers for book CRUD operations
│   │   ├── order_handler.go        # HTTP handlers for order management
│   │   └── user_handler.go         # HTTP handlers for user profile inspection
│   ├── middleware/
│   │   ├── auth.go                 # JWT authentication & RequireRole RBAC middleware
│   │   └── logging_recovery_cors.go# Global request logger, panic recovery, & CORS chain
│   ├── model/
│   │   ├── book.go                 # Book domain model (CanFulfill, Version, Soft-delete)
│   │   ├── order.go                # Order domain model & status definitions
│   │   └── user.go                 # User domain model (RoleAdmin, RoleCustomer constants)
│   ├── db/
│   │   └── mongo.go                # MongoDB connection pool initialization
│   ├── repository/
│   │   ├── book_repository.go      # BookRepository interface
│   │   ├── json_book_repository.go # JSON file storage (sync.RWMutex)
│   │   ├── mongo_book_repository.go# MongoDB storage with indexes
│   │   ├── order_repository.go     # OrderRepository interface
│   │   ├── json_order_repository.go# JSON order repo (injects BookRepo)
│   │   ├── mongo_order_repository.go# MongoDB order repo (injects BookRepo)
│   │   ├── user_repository.go      # UserRepository interface
│   │   ├── json_user_repository.go # JSON user storage
│   │   └── mongo_user_repository.go# MongoDB user storage
│   ├── router/
│   │   └── router.go               # gorilla/mux route configuration & RBAC binding
│   └── service/
│       ├── auth_service.go         # Authentication business logic
│       ├── book_service.go         # Book catalog operations & optimistic locking
│       ├── order_service.go        # Order fulfillment & stock validation logic
│       └── user_service.go         # User registration & profile management
├── data/
│   ├── books.json                  # Book storage seed file
│   ├── orders.json                 # Order storage file
│   └── users.json                  # User identity seed file
├── go.mod                          # Go module dependencies
└── README.md                       # Comprehensive documentation

```

---

## Production Edge Cases & Design Rules

* **Fail-Fast Environment Configuration**: The server checks for `JWT_SECRET` on startup and aborts execution via `log.Fatal` if unset, preventing accidental deployment with hardcoded fallback keys.
* **Global Panic Recovery**: Unhandled runtime panics are intercepted by `RecoveryMiddleware`, returning an HTTP `500 Internal Server Error` without crashing the main process.
* **Pluggable Storage Architecture**: Repository layer abstracts persistence via Go interfaces (`BookRepository`, `UserRepository`, `OrderRepository`). MongoDB implementation uses native driver with connection pooling (10-100 connections), atomic `FindOneAndUpdate` for CAS operations, and case-insensitive unique indexes. JSON implementation uses file-specific `sync.RWMutex` for thread safety.
* **Multi-Phase Stock Deductions**:
  * **Phase 1 (Validation)**: Iterates over all requested order items, checks stock availability in memory, and validates request shapes without writing to storage.
  * **Phase 2 (Mutation)**: Commits stock deductions and order status updates atomically. MongoDB uses `FindOneAndUpdate` with status filters; JSON uses mutex-guarded two-pass validation.


* **Case-Insensitive Duplicate Prevention**: Utilizes `strings.EqualFold()` to prevent duplicate Book entries (Title + Author) and duplicate user registrations (Email).
* **Optimistic Locking**: Implements version-based concurrency control on books (`Version *int`). Updates (`PUT`) verify the client's version token matches current storage before incrementing (`version++`).
* **Soft Deletes**: Deletions mark `DeletedAt *time.Time` instead of purging records, retaining historical audit trails while excluding deleted items from read operations.
* **Security Context Isolation**: Client-submitted order payloads (`CreateOrderRequestDTO`) intentionally exclude unit prices and user IDs. Unit prices are derived server-side from `BookRepository`, and caller identity is extracted from verified JWT claims in context.

---

## HTTP Status Code Matrix

| Status Code | Scenario | Example |
| --- | --- | --- |
| `200 OK` | Successful retrieval or update | `GET /books`, `PUT /books/{id}`, `POST /auth/login` |
| `201 Created` | Successful resource creation | `POST /books`, `POST /users/register`, `POST /orders` |
| `204 No Content` | Successful deletion with no response body | `DELETE /books/{id}` |
| `400 Bad Request` | Structural validation failure or malformed payload | Invalid JSON syntax, missing required payload fields |
| `401 Unauthorized` | Invalid or missing authentication | Missing Authorization header, expired or invalid JWT |
| `403 Forbidden` | Insufficient RBAC privileges or invalid resource ownership | Customer attempting to delete a book or read another user's order |
| `404 Not Found` | Resource missing or soft-deleted | `GET /books/{id}` or `GET /orders/{id}` for non-existent IDs |
| `409 Conflict` | Business rule conflict or version collision | Duplicate user email, duplicate book title, optimistic lock version mismatch, out of stock |
| `500 Internal Server Error` | Storage or server error | Intercepted runtime panics, disk I/O failure |

---

## RBAC & Authentication

### Roles

* **`CUSTOMER`**: Default role for user registration. Granted access to profile inspection, book catalog browsing, order placement, and personal order tracking.
* **`ADMIN`**: Privileged role. Full access to create, update, and soft-delete books, alongside inspecting system-wide orders.

### Access Rules

* **Public**: `POST /auth/login`, `POST /users/register`, `GET /books`, `GET /books/{id}`
* **Protected (Authenticated User)**: `GET /users/me`, `POST /orders`, `GET /orders`, `GET /orders/{id}`
* **Admin Only**: `POST /books`, `PUT /books/{id}`, `DELETE /books/{id}`

---

## Storage Backends

The application supports two storage backends via a pluggable repository pattern, selectable at runtime using the `STORAGE_TYPE` environment variable.

### MongoDB (Recommended for Production)

- **Native Driver**: Uses `go.mongodb.org/mongo-driver` with connection pooling (10-100 connections)
- **Atomic CAS Operations**: `FindOneAndUpdate` with version filters for optimistic locking
- **Case-Insensitive Unique Indexes**: 
  - Books: Compound index on `(title, author)` with collation strength 2
  - Users: Unique index on `email` with case-insensitive matching
- **Partial Indexes**: Excludes soft-deleted documents (`deleted_at: nil`) from uniqueness constraints
- **Graceful Connection Lifecycle**: Context-based timeouts for connect/ping/disconnect operations
- **Repository Implementations**: `mongo_book_repository.go`, `mongo_user_repository.go`, `mongo_order_repository.go`

### JSON Files (Development/Testing)

- **Thread-Safe File I/O**: Dedicated `sync.RWMutex` per file (`books.json`, `users.json`, `orders.json`)
- **In-Memory CAS**: Version comparison under mutex lock before write
- **Suitable For**: Local development, CI/CD testing, demos, and prototyping
- **Scalability Limit**: Not recommended for production workloads exceeding ~1000 records per file
- **Repository Implementations**: `json_book_repository.go`, `json_user_repository.go`, `json_order_repository.go`

---

## Getting Started

### Prerequisites

* **Go 1.22+** installed.

### Setup & Execution

1. **Clone the repository:**
```bash
git clone [https://github.com/TanviShetty19/Book-Store-API.git](https://github.com/TanviShetty19/Book-Store-API.git)
cd Book-Store-API

```


2. **Configure environment variables:**

**Required Environment Variables**

| Variable | Required For | Description | Example |
|----------|--------------|-------------|---------|
| `JWT_SECRET` | All | Secret key for JWT signing/verification. Server refuses to start if unset. | `your-super-secret-key-change-in-prod` |
| `STORAGE_TYPE` | All | Storage backend selector: `mongo` or `json`. Defaults to `mongo`. | `mongo` |
| `MONGO_URI` | MongoDB only | MongoDB connection string with credentials and cluster info. | `mongodb+srv://user:pass@cluster.mongodb.net/` |
| `MONGO_DB_NAME` | MongoDB only | Target database name within the MongoDB cluster. | `bookstore_prod` |
| `PORT` | Optional | HTTP server port. Defaults to `:8080`. | `:3000` |

**Setup Example (MongoDB - Production):**
```bash
export JWT_SECRET="your-super-secret-key-change-in-prod"
export STORAGE_TYPE="mongo"
export MONGO_URI="mongodb+srv://user:pass@cluster.mongodb.net/"
export MONGO_DB_NAME="bookstore_prod"
go run cmd/api/main.go
```

**Setup Example (JSON Files - Development):**
```bash
export JWT_SECRET="dev-secret-key"
export STORAGE_TYPE="json"
# Initialize empty JSON files
echo "[]" > data/users.json && echo "[]" > data/books.json && echo "[]" > data/orders.json
go run cmd/api/main.go
```



---

## API Endpoints

| Method | Endpoint | Description | Auth Required | Expected Status Codes |
| --- | --- | --- | --- | --- |
| `POST` | `/users/register` | Register new user account | No | `201 Created`, `400 Bad Request`, `409 Conflict` |
| `POST` | `/auth/login` | Authenticate user & receive JWT | No | `200 OK`, `400 Bad Request`, `401 Unauthorized` |
| `GET` | `/users/me` | Fetch authenticated user profile | Yes | `200 OK`, `401 Unauthorized` |
| `GET` | `/books` | Retrieve all active books | No | `200 OK` |
| `GET` | `/books/{id}` | Retrieve book by ID | No | `200 OK`, `404 Not Found` |
| `POST` | `/books` | Create a new book record | Admin | `201 Created`, `400 Bad Request`, `403 Forbidden`, `409 Conflict` |
| `PUT` | `/books/{id}` | Update book with optimistic locking | Admin | `200 OK`, `400 Bad Request`, `403 Forbidden`, `404 Not Found`, `409 Conflict` |
| `DELETE` | `/books/{id}` | Soft-delete book record | Admin | `204 No Content`, `403 Forbidden`, `404 Not Found` |
| `POST` | `/orders` | Create order with stock deduction | Yes | `201 Created`, `400 Bad Request`, `401 Unauthorized`, `409 Conflict` |
| `GET` | `/orders` | Retrieve authenticated user's orders | Yes | `200 OK`, `401 Unauthorized` |
| `GET` | `/orders/{id}` | Retrieve order by ID | Yes | `200 OK`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found` |

---

## Sequential Test Suite (All cURL Commands)

Run these commands in order from a second terminal window to execute all operations, edge cases, and verification steps.

### Phase 1: Registration, Login & Profile Operations

```bash
# Register Customer Account (HTTP 201)
curl -i -X POST http://localhost:8080/users/register -H "Content-Type: application/json" -d '{"email":"customer@example.com","password":"password123","role":"CUSTOMER"}'

# Register Admin Account (HTTP 201)
curl -i -X POST http://localhost:8080/users/register -H "Content-Type: application/json" -d '{"email":"admin@example.com","password":"adminpassword123","role":"ADMIN"}'

# Authenticate Customer & Extract Token
CUSTOMER_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '{"email":"customer@example.com","password":"password123"}' \vert{} jq -r '.token') && echo "Customer Token: $CUSTOMER_TOKEN"

# Authenticate Admin & Extract Token
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '{"email":"admin@example.com","password":"adminpassword123"}' \vert{} jq -r '.token') && echo "Admin Token: $ADMIN_TOKEN"

# Get Authenticated User Profile (HTTP 200)
curl -i -X GET http://localhost:8080/users/me -H "Authorization: Bearer $CUSTOMER_TOKEN"

```

---

### Phase 2: Catalog Seeding (Single & Batch Shell Execution)

```bash
# Seed Book 1 (High Stock)
BOOK_1=$(curl -s -X POST http://localhost:8080/books -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" -d '{"title":"Go Design Patterns","author":"Erich Gamma","price":30.00,"stock":50}' \vert{} jq -r '.id') && echo "Book 1 ID: $BOOK_1"

# Seed Book 2 (Low Stock)
BOOK_2=$(curl -s -X POST http://localhost:8080/books -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" -d '{"title":"Clean Architecture in Go","author":"Robert Martin","price":40.00,"stock":2}' \vert{} jq -r '.id') && echo "Book 2 ID: $BOOK_2"

# Batch Creation via Terminal Loop (Adds Book 3 and Book 4)
for book in '{"title":"Concurrency in Go","author":"Katherine Cox-Buday","price":35.00,"stock":10}' '{"title":"The Go Programming Language","author":"Alan Donovan","price":45.00,"stock":15}'; do curl -s -X POST http://localhost:8080/books -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" -d "$book"; done

# Fetch Full Catalog (HTTP 200)
curl -i -X GET http://localhost:8080/books

# Fetch Single Book by ID (HTTP 200)
curl -i -X GET http://localhost:8080/books/$BOOK_1

# Update Book Details (Version 1 -> Version 2) (HTTP 200)
curl -i -X PUT http://localhost:8080/books/$BOOK_1 -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" -d '{"title":"Go Design Patterns (2nd Ed)","price":34.99,"version":1}'

```

---

### Phase 3: Comprehensive Edge Case & Security Validation

```bash
# 1. Server Refuses Startup Without JWT_SECRET
unset JWT_SECRET && go run cmd/api/main.go
# Result: Fatal exit - "JWT_SECRET environment variable is required; server refusing to start without it"

# Restore JWT_SECRET for remaining tests
export JWT_SECRET="your-super-secret-key-change-in-prod"

# 2. Duplicate Email Registration -> Expect HTTP 409 Conflict
curl -i -X POST http://localhost:8080/users/register -H "Content-Type: application/json" -d '{"email":"customer@example.com","password":"password123","role":"CUSTOMER"}'

# 3. Invalid Login Credentials -> Expect HTTP 401 Unauthorized
curl -i -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '{"email":"customer@example.com","password":"wrongpassword"}'

# 4. Non-Admin Creating Book (RBAC Violation) -> Expect HTTP 403 Forbidden
curl -i -X POST http://localhost:8080/books -H "Content-Type: application/json" -H "Authorization: Bearer $CUSTOMER_TOKEN" -d '{"title":"Forbidden Book","price":10.00,"stock":5}'

# 5. Optimistic Concurrency Collision (Stale Version Update) -> Expect HTTP 409 Conflict
curl -i -X PUT http://localhost:8080/books/$BOOK_1 -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" -d '{"price":20.00,"version":1}'

# 6. Multi-Item Order Transaction Failure (Book 2 stock exceeded) -> Expect HTTP 409 Conflict
curl -i -X POST http://localhost:8080/orders -H "Content-Type: application/json" -H "Authorization: Bearer $CUSTOMER_TOKEN" -d '{"items":[{"book_id":"'"$BOOK_1"'","quantity":2},{"book_id":"'"$BOOK_2"'","quantity":10}]}'

# 7. Verify Transaction Atomic Rollback (Book 1 stock must remain untouched at 50)
curl -s http://localhost:8080/books/$BOOK_1 | jq -r '.stock'

# 8. Order Non-Existent Book -> Expect HTTP 404 Not Found
curl -i -X POST http://localhost:8080/orders -H "Content-Type: application/json" -H "Authorization: Bearer $CUSTOMER_TOKEN" -d '{"items":[{"book_id":"00000000-0000-0000-0000-000000000000","quantity":1}]}'

```

---

### Phase 4: Order Execution & Inventory Management

```bash
# Valid Multi-Item Order Execution (Deducts stock for Book 1 and Book 2) -> Expect HTTP 201 Created
ORDER_ID=$(curl -s -X POST http://localhost:8080/orders -H "Content-Type: application/json" -H "Authorization: Bearer $CUSTOMER_TOKEN" -d '{"items":[{"book_id":"'"$BOOK_1"'","quantity":2},{"book_id":"'"$BOOK_2"'","quantity":2}]}' \vert{} jq -r '.id') && echo "Order ID: $ORDER_ID"

# Verify Automatic Inventory Deduction (Book 2 stock is now 0)
curl -s http://localhost:8080/books/$BOOK_2 | jq -r '.stock'

# Fetch Customer Orders (HTTP 200)
curl -i -X GET http://localhost:8080/orders -H "Authorization: Bearer $CUSTOMER_TOKEN"

# Fetch Specific Order Details (HTTP 200)
curl -i -X GET http://localhost:8080/orders/$ORDER_ID -H "Authorization: Bearer $CUSTOMER_TOKEN"

# Admin Accessing Order Details (Admin Override) -> Expect HTTP 200 OK
curl -i -X GET http://localhost:8080/orders/$ORDER_ID -H "Authorization: Bearer $ADMIN_TOKEN"

# Order Attempt on Depleted Stock (Book 2 stock is 0) -> Expect HTTP 409 Conflict
curl -i -X POST http://localhost:8080/orders -H "Content-Type: application/json" -H "Authorization: Bearer $CUSTOMER_TOKEN" -d '{"items":[{"book_id":"'"$BOOK_2"'","quantity":1}]}'

```

---

### Phase 5: Deletion & Soft-Delete Catalog Verification

```bash
# Soft Delete Book (Admin Only) -> Expect HTTP 204 No Content
curl -i -X DELETE http://localhost:8080/books/$BOOK_1 -H "Authorization: Bearer $ADMIN_TOKEN"

# Query Soft-Deleted Book by ID -> Expect HTTP 404 Not Found
curl -i -X GET http://localhost:8080/books/$BOOK_1

# Soft Delete Multiple Books via Shell Execution -> Expect HTTP 204 No Content
for id in $BOOK_2; do curl -i -X DELETE http://localhost:8080/books/$id -H "Authorization: Bearer $ADMIN_TOKEN"; done

```

---

### 1. Get All Orders for the Authenticated User (Customer or Admin)

Sends the JWT token in the `Authorization` header to fetch orders tied to that account:

```bash
curl -i -X GET http://localhost:8080/orders \
  -H "Authorization: Bearer $CUSTOMER_TOKEN"

```

---

### 2. Get a Specific Order by ID

Fetches full item details and status for a single order record using its UUID:

```bash
curl -i -X GET http://localhost:8080/orders/YOUR_ORDER_ID_HERE \
  -H "Authorization: Bearer $CUSTOMER_TOKEN"

```

---

### 3. One-Liner (Login + Fetch Orders Immediately)

If you don't have `$CUSTOMER_TOKEN` saved in your environment variables, run this combined one-liner to log in and immediately return your orders:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '{"email":"customer@example.com","password":"password123"}' | jq -r '.token') && curl -i -X GET http://localhost:8080/orders -H "Authorization: Bearer $TOKEN"

```

---

## MongoDB Inspection & Verification (mongosh)

If running with `STORAGE_TYPE=mongo`, use these commands to inspect database state directly:

```bash
# Connect to your MongoDB cluster
mongosh "mongodb+srv://user:pass@cluster.mongodb.net/bookstore_prod"

# List all collections
show collections

# Verify indexes on books collection (should show unique compound index on title+author)
db.books.getIndexes()

# Count active (non-deleted) books
db.books.countDocuments({ deleted_at: null })

# Inspect a specific book by ID
db.books.findOne({ _id: "BOOK_ID_HERE" })

# Verify version increments after stock deduction
db.books.find({ title: "Go Design Patterns" }, { title: 1, stock: 1, version: 1 })

# List all orders for a specific user
db.orders.find({ user_id: "USER_ID_HERE" })

# Check order status distribution
db.orders.aggregate([
  { $group: { _id: "$status", count: { $sum: 1 } } }
])

# Verify unique email constraint on users collection
db.users.getIndexes()
db.users.findOne({ email: /customer@example.com/i })

# Check connection pool stats (requires admin privileges)
db.serverStatus().connections
```

---

## Architectural Limits & Storage Tradeoffs

### MongoDB Backend

- **Connection Pool Overhead:** Configured with 10-100 connection pool size. Under extreme load, pool exhaustion may cause request queueing (mitigated via `SetMaxConnIdleTime` and `SetServerSelectionTimeout`).
- **Network Latency:** Remote MongoDB clusters introduce network round-trip overhead compared to local file I/O. Partially offset by connection pooling and batch operations.
- **Dependency:** Requires external MongoDB cluster availability. Outages block all persistence operations (no local fallback).

### JSON Backend

- **$O(N)$ File I/O Ceiling:** Each operation reads/writes complete JSON arrays. Thread-safe via `sync.RWMutex`, but high-throughput writes present $O(N)$ overhead.
- **Non-Transactional Multi-File Writes:** Lacks database Write-Ahead Logs (WAL). Multi-file atomicity (`books.json` + `orders.json`) cannot withstand unexpected OS power failures mid-write.
- **Scalability Limit:** Suitable for development/testing only. Not recommended for production workloads exceeding ~1000 records per file.

**Recommendation:** Use MongoDB (`STORAGE_TYPE=mongo`) for production deployments. Use JSON (`STORAGE_TYPE=json`) for local development, CI/CD testing, and demos.

---

## Dependencies

Core Go packages:
- `github.com/gorilla/mux` - HTTP routing and middleware
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/google/uuid` - UUID generation
- `github.com/joho/godotenv` - Environment variable loading from `.env` files
- `go.mongodb.org/mongo-driver` - Official MongoDB driver with connection pooling

Install all dependencies:
```bash
go mod download
```

---

## License

Distributed under the MIT License. See `LICENSE` for more information.

```

```