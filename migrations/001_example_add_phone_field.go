package migrations

import (
	"context"
	"e-document-backend/internal/migration"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Migration001_AddPhoneFieldToUsers adds phone field to all existing users
func Migration001_AddPhoneFieldToUsers() migration.MigrationDefinition {
	return migration.MigrationDefinition{
		Version:     "001",
		Name:        "add_phone_field_to_users",
		Description: "Add phone field to all existing users with default empty string",
		Up:          migration001Up,
		Down:        migration001Down,
	}
}

func migration001Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Update all users that don't have phone field
	filter := bson.M{
		"phone": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"phone": "", // Default empty string
		},
	}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	log.Printf("  Updated %d users with phone field", result.ModifiedCount)
	return nil
}

func migration001Down(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Remove phone field from all users
	filter := bson.M{}
	update := bson.M{
		"$unset": bson.M{
			"phone": "",
		},
	}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	log.Printf("  Removed phone field from %d users", result.ModifiedCount)
	return nil
}
