package repository_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"bookstore-api/internal/apperrors"
	"bookstore-api/internal/db"
	"bookstore-api/internal/model"
	"bookstore-api/internal/repository"

	"go.mongodb.org/mongo-driver/mongo"
)

// setupTestDB connects to a local MongoDB instance using a dedicated test database name.
// It cleans up (drops) the test collection before each test to guarantee complete test isolation.
func setupTestDB(t *testing.T) (*mongo.Client, repository.BookRepository, func()) {
	t.Helper()

	mongoURI := os.Getenv("TEST_MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	testDBName := "bookstore_test_db"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, database, err := db.ConnectMongo(ctx, mongoURI, testDBName)
	if err != nil {
		t.Fatalf("Failed to connect to test MongoDB: %v", err)
	}

	repo := repository.NewMongoBookRepository(database)

	// Teardown function passed back to caller to clean up database state
	teardown := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()

		// Drop the entire test database to leave zero artifacts behind
		_ = database.Drop(cleanupCtx)
		_ = client.Disconnect(cleanupCtx)
	}

	return client, repo, teardown
}

func TestMongoBookRepository_CreateAndGetByID(t *testing.T) {
	_, repo, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()

	newBook := &model.Book{
		ID:        "test_bk_001",
		Title:     "Go Testing in Practice",
		Author:    "Jane Doe",
		Price:     29.99,
		Stock:     10,
		Version:   1,
		CreatedAt: time.Now().Truncate(time.Millisecond),
		UpdatedAt: time.Now().Truncate(time.Millisecond),
	}

	// 1. Test Create
	err := repo.Create(ctx, newBook)
	if err != nil {
		t.Fatalf("Expected no error on book creation, got: %v", err)
	}

	// 2. Test GetByID
	fetchedBook, err := repo.GetByID(ctx, newBook.ID)
	if err != nil {
		t.Fatalf("Expected to find book by ID, got error: %v", err)
	}

	if fetchedBook.Title != newBook.Title {
		t.Errorf("Expected title '%s', got '%s'", newBook.Title, fetchedBook.Title)
	}
	if fetchedBook.Stock != newBook.Stock {
		t.Errorf("Expected stock %d, got %d", newBook.Stock, fetchedBook.Stock)
	}
}

func TestMongoBookRepository_GetByID_NotFound(t *testing.T) {
	_, repo, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()

	_, err := repo.GetByID(ctx, "non_existent_id")
	if err == nil {
		t.Fatal("Expected error when fetching non-existent book ID, got nil")
	}

	// Ensure repository properly maps Mongo ErrNoDocuments to our domain apperrors.ErrNotFound
	if err != apperrors.ErrNotFound {
		t.Errorf("Expected apperrors.ErrNotFound, got: %v", err)
	}
}

func TestMongoBookRepository_SoftDelete(t *testing.T) {
	_, repo, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()

	book := &model.Book{
		ID:      "test_bk_002",
		Title:   "To Be Deleted",
		Author:  "John Smith",
		Price:   15.00,
		Stock:   5,
		Version: 1,
	}

	_ = repo.Create(ctx, book)

	// 1. Perform Soft Delete
	err := repo.Delete(ctx, book.ID)
	if err != nil {
		t.Fatalf("Expected successful delete, got: %v", err)
	}

	// 2. Verify GetByID returns ErrNotFound because IsDeleted is set to true
	_, err = repo.GetByID(ctx, book.ID)
	if err != apperrors.ErrNotFound {
		t.Errorf("Expected soft-deleted book to return ErrNotFound, got: %v", err)
	}
}

func TestMongoBookRepository_OptimisticLocking(t *testing.T) {
	_, repo, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()

	book := &model.Book{
		ID:      "test_bk_003",
		Title:   "Concurrent Book",
		Author:  "Alice",
		Price:   20.00,
		Stock:   10,
		Version: 1, // Initial Version
	}

	_ = repo.Create(ctx, book)

	// 1. Update book with correct version (1) -> Should Succeed & Increment Version to 2
	book.Stock = 8
	err := repo.Update(ctx, book)
	if err != nil {
		t.Fatalf("Expected update with valid version to succeed, got: %v", err)
	}

	// 2. Attempt update using outdated version (Version 1 again instead of 2) -> Should Fail
	outdatedBook := &model.Book{
		ID:      book.ID,
		Title:   "Concurrent Book Modified",
		Stock:   5,
		Version: 1, // Stale version!
	}

	err = repo.Update(ctx, outdatedBook)
	if err == nil {
		t.Fatal("Expected optimistic locking error on stale update, got nil")
	}
	if !errors.Is(err, apperrors.ErrConflict) {
		t.Errorf("Expected apperrors.ErrConflict, got: %v", err)
	}

}
