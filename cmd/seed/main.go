package main

import (
	"context"
	"log"
	"os"
	"time"

	"bookstore-api/internal/db"
	"bookstore-api/internal/model"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found; using system environment variables")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "bookstore"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, database, err := db.ConnectMongo(ctx, mongoURI, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting client: %v", err)
		}
	}()

	log.Println("Connected to MongoDB. Starting database seeding...")

	seedUsers(ctx, database)
	seedBooks(ctx, database)

	log.Println("Database seeding completed successfully!")
}

func seedUsers(ctx context.Context, database *mongo.Database) {
	userCol := database.Collection("users")

	count, err := userCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}
	if count > 0 {
		log.Println("Users collection already contains data. Skipping user seed.")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	adminUser := model.User{
		ID:        "usr_admin01",
		Email:     "admin@bookstore.com",
		Password:  string(hashedPassword),
		Role:      model.RoleAdmin,
		CreatedAt: time.Now(),
	}

	if err := adminUser.Validate(); err != nil {
		log.Fatalf("Invalid admin seed user: %v", err)
	}

	_, err = userCol.InsertOne(ctx, adminUser)
	if err != nil {
		log.Fatalf("Failed to insert admin user: %v", err)
	}

	log.Println("Seeded 1 admin user (Email: 'admin@bookstore.com', Password: 'admin123')")
}

func seedBooks(ctx context.Context, database *mongo.Database) {
	bookCol := database.Collection("books")

	count, err := bookCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to count books: %v", err)
	}
	if count > 0 {
		log.Println("Books collection already contains data. Skipping book seed.")
		return
	}

	now := time.Now()
	sampleBooks := []interface{}{
		model.Book{
			ID:        "bk_001",
			Title:     "The Go Programming Language",
			Author:    "Alan A. A. Donovan & Brian W. Kernighan",
			Price:     39.99,
			Stock:     25,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		model.Book{
			ID:        "bk_002",
			Title:     "Designing Data-Intensive Applications",
			Author:    "Martin Kleppmann",
			Price:     49.99,
			Stock:     15,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		model.Book{
			ID:        "bk_003",
			Title:     "Clean Code",
			Author:    "Robert C. Martin",
			Price:     32.50,
			Stock:     30,
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	_, err = bookCol.InsertMany(ctx, sampleBooks)
	if err != nil {
		log.Fatalf("Failed to insert sample books: %v", err)
	}

	log.Printf("Seeded %d sample books into database.", len(sampleBooks))
}
