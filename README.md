Here is the updated, comprehensive **`README.md`** that merges your existing documentation with all the new features, architecture decisions, and edge-case handling we worked through (including order management, global middleware panic recovery, environment variable safeguards, and multi-phase stock deduction mechanics).

---

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
│  • Thread-safe JSON storage using file-specific sync.RWMutex│
│  • Single-file lock boundaries (BookRepo injected into      │
│    OrderRepo to eliminate dual-lock race conditions)        │
│  • Multi-phase stock validation & deduction                 │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Storage (data/)                                             │
│  • books.json: Catalog with versioning & soft-delete        │
│  • users.json: Account identities & hashed credentials      │
│  • orders.json: Order transactions & item snapshots         │
└─────────────────────────────────────────────────────────────┘

```

---

## Features

* **Clean Architecture:** Strict layer isolation (Handler $\rightarrow$ Service $\rightarrow$ Repository) using Go interfaces and DTO boundaries.
* **Thread-Safe File Persistence:** Concurrent safety using dedicated `sync.RWMutex` instances per file (`books.json`, `users.json`, `orders.json`).
* **Multi-Phase Stock Deduction:** Order processing executes in two distinct phases (Phase 1: In-memory stock validation; Phase 2: Disk mutations) to prevent partial stock loss on validation errors.
* **Global Middleware Pipeline:** Standardized request wrapping using `middleware.Chain` incorporating logging, panic recovery, and CORS headers.
* **Strict Startup Safeguards:** Fail-fast server initialization refusing execution without a configured `JWT_SECRET` environment variable.
* **RESTful CRUD Operations:** Full resource management for users, catalog books, and customer orders.
* **JWT Authentication & RBAC:** Centralized token verification delegating to `auth.ValidateToken()` with role enforcement (`ADMIN` vs. `CUSTOMER`).
* **Optimistic Concurrency Control:** Version-based locking (`Version *int`) to prevent lost update collisions on concurrent writes.

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
│   ├── repository/
│   │   ├── book_repository.go      # Book repository interface
│   │   ├── json_book_repository.go # JSON file storage for books (sync.RWMutex)
│   │   ├── json_order_repository.go# Order repo injecting BookRepo for thread-safe stock deduction
│   │   └── json_user_repository.go # JSON storage for user identities
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
* **Single-Lock File Boundaries**: `JsonOrderRepository` receives `BookRepository` via Dependency Injection instead of opening `books.json` directly. Stock mutations route exclusively through `BookRepository.mu`, eliminating multi-lock race conditions over the same physical file.
* **Multi-Phase Stock Deductions**:
* **Phase 1 (Validation)**: Iterates over all requested order items, checks stock availability in memory, and validates request shapes without writing to disk.
* **Phase 2 (Mutation)**: Commits stock deductions to `books.json` and writes the new order to `orders.json` only after all items pass Phase 1 validation.


* **Case-Insensitive Duplicate Prevention**: Utilizes `strings.EqualFold()` to prevent duplicate Book entries (Title + Author) and duplicate user registrations (Email).
* **Optimistic Locking**: Implements version-based concurrency control on books (`Version *int`). Partial updates (`PUT`/`PATCH`) verify the client's version token matches current storage before incrementing (`version++`).
* **Soft Deletes**: Deletions mark `DeletedAt *time.Time` instead of purging records, retaining historical audit trails while excluding deleted items from read operations.
* **Security Context Isolation**: Client-submitted order payloads (`CreateOrderRequestDTO`) intentionally exclude unit prices and user IDs. Unit prices are derived server-side from `BookRepository`, and caller identity is extracted from verified JWT claims in context.

---

## HTTP Status Code Matrix

| Status Code | Scenario | Example |
| --- | --- | --- |
| `200 OK` | Successful retrieval or update | GET /books, PUT /books/{id}, POST /auth/login |
| `201 Created` | Successful resource creation | POST /books, POST /users/register, POST /orders |
| `400 Bad Request` | Structural validation failure or malformed payload | Invalid JSON, missing required fields, non-UUID path parameters |
| `401 Unauthorized` | Invalid or missing authentication | Missing Authorization header, expired or invalid JWT |
| `403 Forbidden` | Insufficient RBAC privileges or invalid resource ownership | Customer attempting to delete a book or read another user's order |
| `404 Not Found` | Resource missing or soft-deleted | GET /books/{id} or GET /orders/{id} for non-existent IDs |
| `409 Conflict` | Business rule conflict or version collision | Duplicate user registration, duplicate book title, optimistic lock version mismatch |
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

## Getting Started

### Prerequisites

* **Go 1.22+** installed.

### Setup & Execution

1. **Clone the repository:**
```bash
git clone [https://github.com/TanviShetty19/Book-Store-API.git](https://github.com/TanviShetty19/Book-Store-API.git)
cd Book-Store-API

```


2. **Set the mandatory `JWT_SECRET` environment variable:**
```bash
# Linux / macOS
export JWT_SECRET="your-super-secret-key-change-in-prod"

# Windows (Command Prompt)
set JWT_SECRET=your-super-secret-key-change-in-prod

# Windows (PowerShell)
$env:JWT_SECRET="your-super-secret-key-change-in-prod"

```


3. **Start the API server:**
```bash
go run cmd/api/main.go

```


4. **Expected Output:**
```text
Server running on http://localhost:8080

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
| `DELETE` | `/books/{id}` | Soft-delete book record | Admin | `200 OK`, `403 Forbidden`, `404 Not Found` |
| `POST` | `/orders` | Create order with stock deduction | Yes | `201 Created`, `400 Bad Request`, `401 Unauthorized`, `409 Conflict` |
| `GET` | `/orders` | Retrieve authenticated user's orders | Yes | `200 OK`, `401 Unauthorized` |
| `GET` | `/orders/{id}` | Retrieve order by ID | Yes | `200 OK`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found` |

---

## Testing with cURL

### 1. Authentication & Users

**Register a Customer:**

```bash
curl -i -X POST http://localhost:8080/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"customer@example.com","password":"password123","role":"CUSTOMER"}'

```

**Register an Admin:**

```bash
curl -i -X POST http://localhost:8080/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"adminpassword123","role":"ADMIN"}'

```

**Login as Admin & Store Token:**

```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"adminpassword123"}' | jq -r '.token')

```

**Login as Customer & Store Token:**

```bash
CUSTOMER_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"customer@example.com","password":"password123"}' | jq -r '.token')

```

**Fetch Current Profile:**

```bash
curl -i -X GET http://localhost:8080/users/me \
  -H "Authorization: Bearer $CUSTOMER_TOKEN"

```

---

### 2. Catalog & Book Operations

**Create a Book (Admin Only):**

```bash
BOOK_ID=$(curl -s -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"title":"Designing Data-Intensive Applications","author":"Martin Kleppmann","price":45.00,"stock":20}' | jq -r '.id')

```

**Get All Books:**

```bash
curl -i -X GET http://localhost:8080/books

```

**Get Book by ID:**

```bash
curl -i -X GET http://localhost:8080/books/$BOOK_ID

```

**Update Book with Optimistic Locking (`version: 1` -> Success):**

```bash
curl -i -X PUT http://localhost:8080/books/$BOOK_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"title":"Designing Data-Intensive Applications (2nd Ed)","price":49.99,"version":1}'

```

**Soft-Delete Book (Admin Only):**

```bash
curl -i -X DELETE http://localhost:8080/books/$BOOK_ID \
  -H "Authorization: Bearer $ADMIN_TOKEN"

```

---

### 3. Orders & Checkout

**Create an Order (Customer):**

```bash
ORDER_ID=$(curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -d '{
    "items": [
      {
        "book_id": "'"$BOOK_ID"'",
        "quantity": 2
      }
    ]
  }' | jq -r '.id')

```

**Get My Orders (Customer):**

```bash
curl -i -X GET http://localhost:8080/orders \
  -H "Authorization: Bearer $CUSTOMER_TOKEN"

```

**Get Order by ID:**

```bash
curl -i -X GET http://localhost:8080/orders/$ORDER_ID \
  -H "Authorization: Bearer $CUSTOMER_TOKEN"

```

---

### 4. Edge Case & Failure Testing

**Server Refuses to Start Without `JWT_SECRET`:**

```bash
unset JWT_SECRET
go run cmd/api/main.go
# Result: Fatal exit - "JWT_SECRET environment variable is required; server refusing to start without it"

```

**Unauthorized Access (401 Unauthorized):**

```bash
curl -i -X GET http://localhost:8080/users/me

```

**Forbidden Operation (Customer attempting Admin creation -> 403 Forbidden):**

```bash
curl -i -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -d '{"title":"Unauthorized Entry","author":"Hacker","price":10.00,"stock":5}'

```

**Duplicate User Registration (409 Conflict):**

```bash
curl -i -X POST http://localhost:8080/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"customer@example.com","password":"password123"}'

```

**Optimistic Concurrency Control Collision (Stale Version -> 409 Conflict):**

```bash
curl -i -X PUT http://localhost:8080/books/$BOOK_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"price":29.99,"version":1}'

```

**Insufficient Stock Request (409 Conflict):**

```bash
curl -i -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -d '{"items":[{"book_id":"'"$BOOK_ID"'","quantity":999999}]}'

```

---

## Architectural Limits & Storage Tradeoffs

* **$O(N)$ File I/O Ceiling**: Each repository operation reads and rewrites complete JSON arrays. While thread-safe for local development using `sync.RWMutex`, high-throughput writes present an $O(N)$ overhead.
* **Non-Transactional Multi-File Writes**: In file-based storage systems lacking database Write-Ahead Logs (WAL), multi-file atomicity (`books.json` and `orders.json`) cannot withstand unexpected operating system power failures mid-write. Production environments resolve this by transitioning from JSON repositories to an ACID-compliant relational database (such as PostgreSQL).

---

## License

Distributed under the MIT License. See `LICENSE` for more information.

```

```