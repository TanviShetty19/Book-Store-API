package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ConnectMongo establishes a client connection pool to MongoDB and validates reachability via Ping.
// Returns the client (for lifecycle disconnect management) and the specific database instance.
func ConnectMongo(ctx context.Context, uri, dbName string) (*mongo.Client, *mongo.Database, error) {
	// Configure Client Options & Connection Pool settings
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMinPoolSize(10).                         // Maintain at least 10 idle connections ready in pool
		SetMaxPoolSize(100).                        // Cap max concurrent socket connections at 100
		SetMaxConnIdleTime(5 * time.Minute).        // Close idle connections after 5 minutes
		SetConnectTimeout(10 * time.Second).        // Timeout threshold for initial TCP handshake
		SetServerSelectionTimeout(5 * time.Second) // Max time to discover/select an active primary node

	// Initialize the Client (Does not perform network I/O immediately)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create mongodb client options: %w", err)
	}

	// Verify reachability by sending a Ping command to the Primary node
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		// Cleanup client resources if ping fails to avoid socket leaks
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("failed to ping mongodb cluster at %s: %w", uri, err)
	}

	db := client.Database(dbName)
	return client, db, nil
}