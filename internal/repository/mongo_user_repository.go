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

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database) *MongoUserRepository {
	repo := &MongoUserRepository{
		collection: db.Collection("users"),
	}
	repo.initIndexes(context.Background())
	return repo
}

// initIndexes enforces unique, case-insensitive email constraints in MongoDB.
func (r *MongoUserRepository) initIndexes(ctx context.Context) {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "email", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetCollation(&options.Collation{Locale: "en", Strength: 2}), // Case-insensitive unique check
	}
	_, _ = r.collection.Indexes().CreateOne(ctx, indexModel)
}

// Create inserts a new user record into MongoDB.
func (r *MongoUserRepository) Create(ctx context.Context, user *model.User) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user with this email already exists")
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetByID fetches a user by their primary key ID.
func (r *MongoUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return &user, nil
}

// GetByEmail fetches a user by their email address.
func (r *MongoUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	// Case-insensitive regex/collation search for exact email match
	filter := bson.M{
		"email": bson.M{
			"$regex":   fmt.Sprintf("^%s$", email),
			"$options": "i",
		},
	}

	var user model.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}

	return &user, nil
}