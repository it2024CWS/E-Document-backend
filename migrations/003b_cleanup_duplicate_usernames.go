package migrations

import (
	"context"
	"e-document-backend/internal/migration"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Migration003b_CleanupDuplicateUsernames removes duplicate username entries before creating unique index
func Migration003b_CleanupDuplicateUsernames() migration.MigrationDefinition {
	return migration.MigrationDefinition{
		Version:     "003b",
		Name:        "cleanup_duplicate_usernames",
		Description: "Remove duplicate username entries to prepare for unique index",
		Up:          migration003bUp,
		Down:        migration003bDown,
	}
}

func migration003bUp(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Find all duplicate usernames
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$username",
				"count": bson.M{"$sum": 1},
				"docs":  bson.M{"$push": "$$ROOT"},
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gt": 1},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	duplicatesRemoved := 0
	for cursor.Next(ctx) {
		var result struct {
			ID    string   `bson:"_id"`
			Count int      `bson:"count"`
			Docs  []bson.M `bson:"docs"`
		}

		if err := cursor.Decode(&result); err != nil {
			log.Printf("Error decoding result: %v", err)
			continue
		}

		// Keep the first document, delete the rest
		for i := 1; i < len(result.Docs); i++ {
			docID := result.Docs[i]["_id"]
			_, err := collection.DeleteOne(ctx, bson.M{"_id": docID})
			if err != nil {
				log.Printf("Error deleting duplicate document: %v", err)
				continue
			}
			duplicatesRemoved++
		}
	}

	log.Printf("  Removed %d duplicate username records", duplicatesRemoved)
	return nil
}

func migration003bDown(ctx context.Context, db *mongo.Database) error {
	// Cannot restore deleted documents
	log.Println("  Cannot restore deleted duplicate records")
	return nil
}
