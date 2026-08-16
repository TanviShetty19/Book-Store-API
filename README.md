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
│ Router & Middleware (internal/router)                       │
│  • Matches routes using Go 1.22+ native ServeMux            │
│  • Logs request method, path, status code, and execution time│
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Handler Layer — Carl (internal/handler)                    │
│  • Speaks HTTP/JSON, parses URL parameters & request bodies │
│  • Maps Go errors to HTTP status codes (200, 201, 400, 404) │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Service Layer — Bob (internal/service)                      │
│  • Enforces business validation rules (Title, Author, Price)│
│  • Manages CreatedAt and UpdatedAt timestamps               │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Repository Layer — Sam (internal/repository)                │
│  • Thread-safe JSON file I/O using sync.RWMutex             │
│  • Reads and writes to data/books.json                      │
└─────────────────────────────────────────────────────────────┘

```

---

##  Features

* **Clean Architecture:** Strict layer isolation (Handler $\rightarrow$ Service $\rightarrow$ Repository) using Go interfaces.
* **Thread-Safe Data Persistence:** Concurrent safety using `sync.RWMutex` for atomic reads and writes to `data/books.json`.
* **Zero External Dependencies:** Built entirely with Go standard library (`net/http`, `encoding/json`, `sync`, `time`).
* **HTTP Request Middleware:** Structured access logging capturing execution latency, status codes, and path details.
* **RESTful CRUD Operations:** Full resource management over standard HTTP verbs.

---

##  Project Structure

```text
bookstore-api/
├── cmd/
│   └── api/
│       └── main.go           # Application entrypoint & dependency injection
├── internal/
│   ├── handler/
│   │   └── book_handler.go   # HTTP handlers (Carl)
│   ├── service/
│   │   └── book_service.go   # Business logic & validation (Bob)
│   ├── repository/
│   │   ├── book_repository.go      # Repository interface (Sam's Contract)
│   │   └── json_book_repository.go # Thread-safe JSON storage implementation
│   ├── router/
│   │   └── router.go         # Route registration & logging middleware
│   └── model/
│       └── book.go           # Shared Book domain model struct
├── data/
│   └── books.json            # JSON database seed file
├── go.mod                    # Go module definition
└── README.md                 # Documentation

```

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

| Method | Endpoint | Description | Expected Status |
| --- | --- | --- | --- |
| `GET` | `/books` | Retrieve all books | `200 OK` |
| `GET` | `/books/{id}` | Retrieve a specific book by ID | `200 OK` / `404 Not Found` |
| `POST` | `/books` | Create a new book record | `201 Created` / `400 Bad Request` |
| `PUT` | `/books/{id}` | Update an existing book record | `200 OK` / `400` / `404` |
| `DELETE` | `/books/{id}` | Delete a book record by ID | `204 No Content` / `404` |

---

##  Testing with cURL

Open a second terminal window and execute the following commands to test the endpoints:

### 1. Get All Books

```bash
curl -i http://localhost:8080/books

```

### 2. Get Book by ID

```bash
curl -i http://localhost:8080/books/1

```

### 3. Create a New Book

```bash
curl -i -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Designing Data-Intensive Applications",
    "author": "Martin Kleppmann",
    "price": 45.00
  }'

```

### 4. Update an Existing Book

```bash
curl -i -X PUT http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Go Programming Language (2nd Edition)",
    "author": "Alan A. A. Donovan",
    "price": 49.99
  }'

```

### 5. Delete a Book

```bash
curl -i -X DELETE http://localhost:8080/books/1

```

---

## License

Distributed under the MIT License. See `LICENSE` for more information.
