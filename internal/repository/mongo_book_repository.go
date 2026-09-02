package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"bookstore-api/internal/apperrors"
	"bookstore-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoBookRepository struct {
	collection *mongo.Collection
}

func NewMongoBookRepository(db *mongo.Database) *MongoBookRepository {
	repo := &MongoBookRepository{
		collection: db.Collection("books"),
	}
	repo.initIndexes(context.Background())
	return repo
}

func (r *MongoBookRepository) initIndexes(ctx context.Context) {
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: 1},
			{Key: "author", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetCollation(&options.Collation{Locale: "en", Strength: 2}).
			SetPartialFilterExpression(bson.M{"deleted_at": nil}),
	}
	_, _ = r.collection.Indexes().CreateOne(ctx, indexModel)
}

func (r *MongoBookRepository) Create(ctx context.Context, book *model.Book) error {
	now := time.Now().UTC()
	book.CreatedAt = now
	book.UpdatedAt = now
	if book.Version == 0 {
		book.Version = 1
	}
	book.DeletedAt = nil

	_, err := r.collection.InsertOne(ctx, book)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("failed to insert book: %w", err)
	}
	return nil
}

func (r *MongoBookRepository) GetByID(ctx context.Context, id string) (*model.Book, error) {
	filter := bson.M{
		"_id":        id,
		"deleted_at": nil,
	}

	var book model.Book
	err := r.collection.FindOne(ctx, filter).Decode(&book)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch book: %w", err)
	}

	return &book, nil
}

// GetAll aligns with JsonBookRepository signature: ([]*model.Book, error)
func (r *MongoBookRepository) GetAll(ctx context.Context) ([]*model.Book, error) {
	filter := bson.M{"deleted_at": nil}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query books: %w", err)
	}
	defer cursor.Close(ctx)

	var activeBooks []*model.Book
	for cursor.Next(ctx) {
		var book model.Book
		if err := cursor.Decode(&book); err != nil {
			return nil, fmt.Errorf("failed to decode book: %w", err)
		}
		activeBooks = append(activeBooks, &book)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return activeBooks, nil
}

// Update performs an atomic compare-and-swap on MongoDB matching JSON contract.
func (r *MongoBookRepository) Update(ctx context.Context, updatedBook *model.Book) error {
	filter := bson.M{
		"_id":        updatedBook.ID,  //Book ID of the book you want to update
		"version":    updatedBook.Version, // Expected current version before increment
		"deleted_at": nil,
	}

	now := time.Now().UTC()
	newVersion := updatedBook.Version + 1

	update := bson.M{
		"$set": bson.M{
			"title":      updatedBook.Title,
			"author":     updatedBook.Author,
			"price":      updatedBook.Price,
			"stock":      updatedBook.Stock,
			"version":    newVersion,
			"updated_at": now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("failed to update book: %w", err)
	}

	if result.MatchedCount == 0 {
		// Verify if document exists to distinguish NotFound vs Version Mismatch (Conflict)
		existing, _ := r.GetByID(ctx, updatedBook.ID)
		if existing == nil {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("%w: book version mismatch (current: %d, provided: %d)", apperrors.ErrConflict, existing.Version, updatedBook.Version)
	}

	updatedBook.Version = newVersion
	updatedBook.UpdatedAt = now
	return nil
}

func (r *MongoBookRepository) Delete(ctx context.Context, id string) error {
	filter := bson.M{
		"_id":        id,
		"deleted_at": nil,
	}

	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to soft delete book: %w", err)
	}

	if result.MatchedCount == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}