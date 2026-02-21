package migration

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MigrationFunc is the function signature for migration up/down operations.
type MigrationFunc func(ctx context.Context, db *mongo.Database) error

// MigrationDefinition describes a single migration step.
type MigrationDefinition struct {
	Version     string
	Name        string
	Description string
	Up          MigrationFunc
	Down        MigrationFunc
}

// migrationRecord represents a migration record stored in MongoDB.
type migrationRecord struct {
	Version   string    `bson:"version"`
	Name      string    `bson:"name"`
	AppliedAt time.Time `bson:"applied_at"`
}

// Runner manages and executes migrations.
type Runner struct {
	db         *mongo.Database
	collection *mongo.Collection
	migrations []MigrationDefinition
}

// NewRunner creates a new migration Runner for the given MongoDB database.
func NewRunner(db *mongo.Database) *Runner {
	return &Runner{
		db:         db,
		collection: db.Collection("_migrations"),
	}
}

// Register adds a migration definition to the runner.
func (r *Runner) Register(def MigrationDefinition) {
	r.migrations = append(r.migrations, def)
}

// RunUp applies all migrations that have not yet been applied.
func (r *Runner) RunUp(ctx context.Context) error {
	if err := r.ensureIndex(ctx); err != nil {
		return fmt.Errorf("migration: failed to create index: %w", err)
	}

	applied, err := r.appliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("migration: failed to load applied versions: %w", err)
	}

	for _, m := range r.migrations {
		if applied[m.Version] {
			log.Printf("migration: [%s] %s — already applied, skipping", m.Version, m.Name)
			continue
		}

		log.Printf("migration: [%s] %s — running UP...", m.Version, m.Name)
		if err := m.Up(ctx, r.db); err != nil {
			return fmt.Errorf("migration: [%s] %s UP failed: %w", m.Version, m.Name, err)
		}

		record := migrationRecord{
			Version:   m.Version,
			Name:      m.Name,
			AppliedAt: time.Now().UTC(),
		}
		if _, err := r.collection.InsertOne(ctx, record); err != nil {
			return fmt.Errorf("migration: failed to record [%s]: %w", m.Version, err)
		}

		log.Printf("migration: [%s] %s — done ✅", m.Version, m.Name)
	}

	return nil
}

// RunDown rolls back all applied migrations in reverse order.
func (r *Runner) RunDown(ctx context.Context) error {
	applied, err := r.appliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("migration: failed to load applied versions: %w", err)
	}

	// Iterate in reverse
	for i := len(r.migrations) - 1; i >= 0; i-- {
		m := r.migrations[i]
		if !applied[m.Version] {
			continue
		}

		log.Printf("migration: [%s] %s — running DOWN...", m.Version, m.Name)
		if m.Down != nil {
			if err := m.Down(ctx, r.db); err != nil {
				return fmt.Errorf("migration: [%s] %s DOWN failed: %w", m.Version, m.Name, err)
			}
		}

		if _, err := r.collection.DeleteOne(ctx, bson.M{"version": m.Version}); err != nil {
			return fmt.Errorf("migration: failed to remove record [%s]: %w", m.Version, err)
		}

		log.Printf("migration: [%s] %s — rolled back ↩️", m.Version, m.Name)
	}

	return nil
}

// appliedVersions returns a set of all already-applied migration versions.
func (r *Runner) appliedVersions(ctx context.Context) (map[string]bool, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	applied := make(map[string]bool)
	for cursor.Next(ctx) {
		var rec migrationRecord
		if err := cursor.Decode(&rec); err != nil {
			return nil, err
		}
		applied[rec.Version] = true
	}
	return applied, cursor.Err()
}

// ensureIndex creates a unique index on the version field in the migrations collection.
func (r *Runner) ensureIndex(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("version_unique_idx"),
	}
	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	return err
}
