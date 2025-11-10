package migrations

import (
	"context"
	"e-document-backend/internal/migration"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Migration002_CreateEmailIndex creates a unique index on email field
func Migration002_CreateEmailIndex() migration.MigrationDefinition {
	return migration.MigrationDefinition{
		Version:     "002",
		Name:        "create_email_index",
		Description: "Create unique index on email field for better query performance",
		Up:          migration002Up,
		Down:        migration002Down,
	}
}

func migration002Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Create unique index on email
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("email_unique_idx"),
	}

	indexName, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	log.Printf("  Created index: %s", indexName)
	return nil
}

func migration002Down(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Drop the email index
	_, err := collection.Indexes().DropOne(ctx, "email_unique_idx")
	if err != nil {
		return err
	}

	log.Println("  Dropped index: email_unique_idx")
	return nil
}
