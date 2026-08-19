# Bookstore REST API

A lightweight, high-performance RESTful API for managing a bookstore catalog built with **Go 1.22+**. Designed following **Clean Architecture (3-Tier Layered Architecture)** principles, featuring thread-safe file persistence and custom HTTP logging middleware.

---

##  Architecture Overview

The application strictly separates concerns into decoupled layers using Go interfaces and Dependency Injection (DI):

```text
[ Client / cURL ]
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│ Middleware Layer (internal/middleware)                      │
│  • JWT Authentication: Validates Bearer tokens               │
│  • RBAC Authorization: Enforces role-based access control    │
│  • Injects user_id and role into request context            │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Handler Layer (internal/handler)                            │
│  • HTTP request/response handling                            │
│  • DTO validation and parsing                               │
│  • Maps service errors to HTTP status codes                  │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Service Layer (internal/service)                            │
│  • Business logic and validation                            │
│  • Optimistic locking and duplicate detection               │
│  • Soft delete management                                    │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Repository Layer (internal/repository)                      │
│  • Data persistence abstraction                              │
│  • Thread-safe JSON file I/O using sync.RWMutex             │
│  • CRUD operations on data/books.json and in-memory users    │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Storage (data/)                                              │
│  • books.json: Book catalog with soft delete support       │
│  • In-memory: User repository for authentication            │
└─────────────────────────────────────────────────────────────┘

```

---

##  Features

* **Clean Architecture:** Strict layer isolation (Handler $\rightarrow$ Service $\rightarrow$ Repository) using Go interfaces.
* **Thread-Safe Data Persistence:** Concurrent safety using `sync.RWMutex` for atomic reads and writes to `data/books.json`.
* **Zero External Dependencies:** Built entirely with Go standard library (`net/http`, `encoding/json`, `sync`, `time`).
* **HTTP Request Middleware:** Structured access logging capturing execution latency, status codes, and path details.
* **RESTful CRUD Operations:** Full resource management over standard HTTP verbs.
* **JWT Authentication & RBAC:** Token-based authentication with role-based access control (admin/customer).
* **Optimistic Locking:** Version-based concurrency control to prevent write collisions.

---

##  Project Structure

```text
bookstore-api/
├── cmd/
│   └── api/
│       └── main.go           # Application entrypoint & dependency injection
├── internal/
│   ├── handler/
│   │   ├── book_handler.go   # HTTP handlers for book CRUD operations
│   │   └── auth_handler.go   # HTTP handlers for authentication
│   ├── service/
│   │   ├── book_service.go   # Business logic for books
│   │   └── auth_service.go   # Business logic for authentication
│   ├── repository/
│   │   ├── book_repository.go      # Repository interface
│   │   ├── json_book_repository.go # JSON file storage implementation
│   │   └── user_repository.go      # User repository interface & implementation
│   ├── dto/
│   │   ├── book_dto.go     # Data Transfer Objects for books
│   │   └── auth_dto.go     # Data Transfer Objects for authentication
│   ├── model/
│   │   ├── book.go         # Book domain model
│   │   └── user.go         # User domain model
│   ├── middleware/
│   │   └── auth_middleware.go  # JWT authentication & RBAC
│   ├── auth/
│   │   └── jwt.go          # JWT token generation and validation
│   └── router/
│       └── router.go       # Route registration
├── data/
│   └── books.json          # JSON database seed file
├── go.mod                  # Go module definition
└── README.md               # Documentation

```

### Directory Responsibilities

- **cmd/api**: Application entrypoint, dependency injection setup, and server initialization
- **internal/model**: Domain entities (Book, User) with business-critical fields
- **internal/dto**: Data Transfer Objects for request validation and response formatting
- **internal/handler**: HTTP request/response handling, status code mapping
- **internal/service**: Business logic, validation, optimistic locking, soft deletes
- **internal/repository**: Data persistence abstraction with thread-safe implementations
- **internal/middleware**: JWT authentication and RBAC authorization
- **internal/auth**: JWT token generation and validation utilities

---

##  Production Edge Cases

The API implements comprehensive edge case handling to ensure robustness in production environments:

- **Whitespace Sanitization & Empty Guard**: Trims leading/trailing whitespace from string inputs (title, author, email) and blocks whitespace-only submissions to prevent database pollution.

- **String Length Limits**: Enforces maximum 255 character limits for titles and authors to prevent storage overflow and maintain consistent data quality.

- **Financial Range Boundaries**: Restricts price values between $0.01 and $10,000.00 to prevent invalid monetary data and potential business logic errors.

- **Case-Insensitive Duplicate Prevention**: Uses `strings.EqualFold()` for case-insensitive matching of Title + Author combinations to prevent duplicate book entries regardless of capitalization.

- **Strict UUID Validation**: Validates all UUID path parameters and request body fields using `uuid.Parse()` to reject malformed identifiers before processing.

- **Optimistic Locking**: Implements version-based concurrency control where each book has a `version` field. Update requests must include the current version; mismatches return `412 Precondition Failed` to prevent lost updates from concurrent writes.

- **No-Op Update Optimization**: Detects when update payloads contain identical data to the existing record and bypasses unnecessary disk writes and timestamp updates, improving performance.

- **Query Pagination**: Caps collection responses with configurable `page` and `limit` parameters (default limit: 20, max: 100) and returns structured metadata including current page, page size, total items, and total pages.

- **Soft Deletes**: Marks records as deleted using `deleted_at` timestamps instead of permanent removal, preserving historical audit data while filtering soft-deleted records from read operations.

---

##  HTTP Status Code Matrix

The API follows REST conventions with precise status code mapping:

| Status Code | Scenario | Example |
| --- | --- | --- |
| `200 OK` | Standard success response | Successful GET requests, successful updates |
| `201 Created` | Resource successfully created | POST /books, POST /auth/register |
| `204 No Content` | Successful operation with no response body | DELETE /books/{id}, DELETE /books/batch |
| `400 Bad Request` | Validation failure or invalid input | Missing required fields, invalid UUID format, price out of range |
| `401 Unauthorized` | Missing or invalid authentication token | No Authorization header, expired JWT |
| `403 Forbidden` | Insufficient role permissions | Customer attempting DELETE /books/{id} |
| `404 Not Found` | Resource missing or soft-deleted | GET /books/{id} for non-existent or deleted book |
| `409 Conflict` | Duplicate resource creation | Duplicate email registration, duplicate book title/author |
| `412 Precondition Failed` | Optimistic locking collision | Update with mismatched version number |
| `500 Internal Server Error` | Unhandled server or storage failure | Database errors, token generation failures |

---

##  RBAC & Authentication

The API implements JWT-based authentication with Role-Based Access Control (RBAC) to secure endpoints.

### User Roles

- **customer**: Default role for new registrations. Can read books and create/update books (with authentication).
- **admin**: Administrative role with additional permissions including delete operations.

### Authentication Flow

1. **Registration**: POST /auth/register with email, password, and optional role
2. **Login**: POST /auth/login with email and password
3. **Token Usage**: Include JWT in Authorization header: `Authorization: Bearer <token>`

### Protected Endpoints

- **Authentication Required**: POST /books, PUT /books/{id}, POST /books/batch
- **Admin Only**: DELETE /books/{id}, DELETE /books/batch

### JWT Token Details

- **Algorithm**: HS256
- **Expiration**: 24 hours from issuance
- **Claims**: user_id, role, issued_at, expires_at

---

##  Getting Started

### Prerequisites

* **Go 1.22** or higher installed on your system.

### Installation & Execution

1. **Clone the repository:**
```bash
git clone https://github.com/TanviShetty19/Book-Store-API.git
cd Book-Store-API

```


2. **Run the server:**
```bash
go run cmd/api/main.go

```


3. **Expected Terminal Output:**
```text
Bookstore API server running on http://localhost:8080...

```



---

## API Endpoints

| Method | Endpoint | Description | Auth Required | Expected Status |
| --- | --- | --- | --- | --- |
| `POST` | `/auth/register` | Register new user | No | `201 Created` / `400 Bad Request` / `409 Conflict` |
| `POST` | `/auth/login` | Login and receive JWT | No | `200 OK` / `401 Unauthorized` |
| `GET` | `/books` | Retrieve all books (paginated) | No | `200 OK` |
| `GET` | `/books/{id}` | Retrieve a specific book by ID | No | `200 OK` / `404 Not Found` |
| `POST` | `/books` | Create a new book record | Yes | `201 Created` / `400 Bad Request` / `409 Conflict` |
| `POST` | `/books/batch` | Create multiple books | Yes | `201 Created` / `400 Bad Request` / `409 Conflict` |
| `PUT` | `/books/{id}` | Update an existing book record | Yes | `200 OK` / `400 Bad Request` / `404 Not Found` / `409 Conflict` / `412 Precondition Failed` |
| `DELETE` | `/books/{id}` | Delete a book record by ID | Admin | `204 No Content` / `403 Forbidden` / `404 Not Found` |
| `DELETE` | `/books/batch` | Delete multiple books | Admin | `204 No Content` / `403 Forbidden` / `404 Not Found` |

---

##  Testing with cURL

### Authentication

**Register as customer:**
```bash
curl -i -X POST http://localhost:8080/auth/register -H "Content-Type: application/json" -d '{"email":"customer@example.com","password":"password123"}'
```

**Register as admin:**
```bash
curl -i -X POST http://localhost:8080/auth/register -H "Content-Type: application/json" -d '{"email":"admin@example.com","password":"admin123","role":"admin"}'
```

**Login and extract token:**
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '{"email":"admin@example.com","password":"admin123"}' | jq -r '.token')
```

### CRUD Operations

**Create a book (requires auth):**
```bash
curl -i -X POST http://localhost:8080/books -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"title":"Clean Code","author":"Robert C. Martin","price":35.00}'
```

**Get all books (paginated):**
```bash
curl -i "http://localhost:8080/books?page=1&limit=2"
```

**Get book by ID:**
```bash
curl -i http://localhost:8080/books/{book_id}
```

**Update book with optimistic locking:**
```bash
curl -i -X PUT http://localhost:8080/books/{book_id} -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"title":"Clean Code (Updated)","author":"Robert C. Martin","price":39.99,"version":1}'
```

**Delete book (admin only):**
```bash
curl -i -X DELETE http://localhost:8080/books/{book_id} -H "Authorization: Bearer $TOKEN"
```

### Edge Case Testing

**Validation error (400):**
```bash
curl -i -X POST http://localhost:8080/books -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"title":"","author":"Test","price":35.00}'
```

**Duplicate conflict (409):**
```bash
curl -i -X POST http://localhost:8080/books -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"title":"Clean Code","author":"Robert C. Martin","price":35.00}'
```

**Optimistic locking failure (412):**
```bash
curl -i -X PUT http://localhost:8080/books/{book_id} -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"title":"Test","author":"Test","price":10.00,"version":99}'
```

**Forbidden - customer attempting delete (403):**
```bash
CUSTOMER_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '{"email":"customer@example.com","password":"password123"}' | jq -r '.token')
curl -i -X DELETE http://localhost:8080/books/{book_id} -H "Authorization: Bearer $CUSTOMER_TOKEN"
```

---

## License

Distributed under the MIT License. See `LICENSE` for more information.
