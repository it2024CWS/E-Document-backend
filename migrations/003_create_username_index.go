package migrations

import (
	"context"
	"e-document-backend/internal/migration"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Migration003_CreateUsernameIndex creates a unique index on username field
func Migration003_CreateUsernameIndex() migration.MigrationDefinition {
	return migration.MigrationDefinition{
		Version:     "003",
		Name:        "create_username_index",
		Description: "Create unique index on username field to prevent duplicates",
		Up:          migration003Up,
		Down:        migration003Down,
	}
}

func migration003Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Create unique index on username
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("username_unique_idx"),
	}

	indexName, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	log.Printf("  Created index: %s", indexName)
	return nil
}

func migration003Down(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Drop the username index
	_, err := collection.Indexes().DropOne(ctx, "username_unique_idx")
	if err != nil {
		return err
	}

	log.Println("  Dropped index: username_unique_idx")
	return nil
}
