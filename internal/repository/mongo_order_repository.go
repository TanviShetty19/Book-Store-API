package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"bookstore-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoOrderRepository struct {
	collection *mongo.Collection
	bookRepo   BookRepository
}

func NewMongoOrderRepository(db *mongo.Database, bookRepo BookRepository) *MongoOrderRepository {
	repo := &MongoOrderRepository{
		collection: db.Collection("orders"),
		bookRepo:   bookRepo,
	}
	repo.initIndexes(context.Background())
	return repo
}

func (r *MongoOrderRepository) initIndexes(ctx context.Context) {
	// Index on user_id for fast order lookups by user
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}
	_, _ = r.collection.Indexes().CreateOne(ctx, indexModel)
}

// CreateDraftOrder persists a new order in PENDING status.
func (r *MongoOrderRepository) CreateDraftOrder(ctx context.Context, order *model.Order) error {
	now := time.Now().UTC()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	order.UpdatedAt = now
	order.Status = model.OrderStatusPending

	_, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}
	return nil
}

// ConfirmOrder performs the authoritative transition from PENDING to CONFIRMED.
// Validates stock availability, deducts stock via bookRepo, and updates order status.
func (r *MongoOrderRepository) ConfirmOrder(ctx context.Context, orderID string) (*model.Order, error) {
	order, err := r.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.Status != model.OrderStatusPending {
		return nil, fmt.Errorf("order is not pending (current status: %s)", order.Status)
	}

	// Pass 1: validate stock availability for every item before mutating anything.
	books := make(map[string]*model.Book, len(order.Items))
	for _, item := range order.Items {
		book, err := r.bookRepo.GetByID(ctx, item.BookID)
		if err != nil {
			return nil, fmt.Errorf("book with ID %s not found: %w", item.BookID, err)
		}
		if !book.CheckStock(item.Quantity) {
			return nil, fmt.Errorf("insufficient stock for '%s': requested %d, available %d", book.Title, item.Quantity, book.Stock)
		}
		books[item.BookID] = book
	}

	// Pass 2: all checks passed, apply deductions via bookRepo.
	for _, item := range order.Items {
		book := books[item.BookID]
		book.Stock -= item.Quantity
		if err := r.bookRepo.Update(ctx, book); err != nil {
			return nil, fmt.Errorf("failed to deduct stock for book '%s': %w", book.Title, err)
		}
	}

	// Pass 3: update order status in MongoDB
	now := time.Now().UTC()
	filter := bson.M{
		"_id":    orderID,
		"status": model.OrderStatusPending,
	}
	update := bson.M{
		"$set": bson.M{
			"status":     model.OrderStatusCompleted,
			"updated_at": now,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var confirmedOrder model.Order
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&confirmedOrder)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("order status changed concurrently or order not found")
		}
		return nil, fmt.Errorf("failed to confirm order: %w", err)
	}

	return &confirmedOrder, nil
}

// CancelOrder transitions a PENDING order to CANCELLED.
func (r *MongoOrderRepository) CancelOrder(ctx context.Context, orderID string) (*model.Order, error) {
	now := time.Now().UTC()
	filter := bson.M{
		"_id":    orderID,
		"status": model.OrderStatusPending,
	}
	update := bson.M{
		"$set": bson.M{
			"status":     model.OrderStatusCancelled,
			"updated_at": now,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var cancelledOrder model.Order
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&cancelledOrder)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Check if order exists at all to return precise error
			existing, getErr := r.GetByID(ctx, orderID)
			if getErr != nil {
				return nil, getErr
			}
			return nil, fmt.Errorf("order is not pending (current status: %s)", existing.Status)
		}
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	return &cancelledOrder, nil
}

// GetByID fetches an order by its primary key ID.
func (r *MongoOrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}

	return &order, nil
}

// GetByUserID fetches all orders associated with a specific user ID.
func (r *MongoOrderRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Order, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to query user orders: %w", err)
	}
	defer cursor.Close(ctx)

	var userOrders []*model.Order
	for cursor.Next(ctx) {
		var order model.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, fmt.Errorf("failed to decode order: %w", err)
		}
		userOrders = append(userOrders, &order)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return userOrders, nil
}
