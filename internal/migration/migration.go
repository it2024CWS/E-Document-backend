package migration

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Migration represents a database migration
type Migration struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Version     string             `bson:"version"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	AppliedAt   time.Time          `bson:"applied_at"`
}

// MigrationFunc is a function that performs a migration
type MigrationFunc func(ctx context.Context, db *mongo.Database) error

// MigrationDefinition defines a migration with up and down functions
type MigrationDefinition struct {
	Version     string
	Name        string
	Description string
	Up          MigrationFunc
	Down        MigrationFunc
}

// Runner manages database migrations
type Runner struct {
	db         *mongo.Database
	collection *mongo.Collection
	migrations []MigrationDefinition
}

// NewRunner creates a new migration runner
func NewRunner(db *mongo.Database) *Runner {
	return &Runner{
		db:         db,
		collection: db.Collection("migrations"),
		migrations: []MigrationDefinition{},
	}
}

// Register adds a migration to the runner
func (r *Runner) Register(migration MigrationDefinition) {
	r.migrations = append(r.migrations, migration)
}

// Up runs all pending migrations
func (r *Runner) Up(ctx context.Context) error {
	log.Println("Running migrations...")

	// Create migrations collection if not exists
	if err := r.ensureMigrationsCollection(ctx); err != nil {
		return err
	}

	// Get applied migrations
	appliedMigrations, err := r.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	appliedMap := make(map[string]bool)
	for _, m := range appliedMigrations {
		appliedMap[m.Version] = true
	}

	// Run pending migrations
	for _, migration := range r.migrations {
		if appliedMap[migration.Version] {
			log.Printf("⊗ Migration %s (%s) already applied", migration.Version, migration.Name)
			continue
		}

		log.Printf("⇒ Running migration %s: %s", migration.Version, migration.Name)

		if err := migration.Up(ctx, r.db); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
		}

		// Record migration
		if err := r.recordMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		log.Printf("✓ Migration %s completed successfully", migration.Version)
	}

	log.Println("All migrations completed!")
	return nil
}

// Down rolls back the last migration
func (r *Runner) Down(ctx context.Context) error {
	log.Println("Rolling back last migration...")

	// Get last applied migration
	lastMigration, err := r.getLastMigration(ctx)
	if err != nil {
		return err
	}

	if lastMigration == nil {
		log.Println("No migrations to roll back")
		return nil
	}

	// Find migration definition
	var migrationDef *MigrationDefinition
	for _, m := range r.migrations {
		if m.Version == lastMigration.Version {
			migrationDef = &m
			break
		}
	}

	if migrationDef == nil {
		return fmt.Errorf("migration definition not found for version %s", lastMigration.Version)
	}

	if migrationDef.Down == nil {
		return fmt.Errorf("no down function defined for migration %s", lastMigration.Version)
	}

	log.Printf("⇐ Rolling back migration %s: %s", migrationDef.Version, migrationDef.Name)

	// Run down migration
	if err := migrationDef.Down(ctx, r.db); err != nil {
		return fmt.Errorf("failed to roll back migration %s: %w", migrationDef.Version, err)
	}

	// Remove migration record
	if err := r.removeMigration(ctx, lastMigration.Version); err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", migrationDef.Version, err)
	}

	log.Printf("✓ Migration %s rolled back successfully", migrationDef.Version)
	return nil
}

// Status shows migration status
func (r *Runner) Status(ctx context.Context) error {
	appliedMigrations, err := r.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	appliedMap := make(map[string]*Migration)
	for i := range appliedMigrations {
		appliedMap[appliedMigrations[i].Version] = &appliedMigrations[i]
	}

	log.Println("Migration Status:")
	log.Println("================")

	for _, migration := range r.migrations {
		if applied, ok := appliedMap[migration.Version]; ok {
			log.Printf("✓ %s - %s (Applied: %s)", migration.Version, migration.Name, applied.AppliedAt.Format(time.RFC3339))
		} else {
			log.Printf("⊗ %s - %s (Pending)", migration.Version, migration.Name)
		}
	}

	return nil
}

// ensureMigrationsCollection creates migrations collection if it doesn't exist
func (r *Runner) ensureMigrationsCollection(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// getAppliedMigrations returns all applied migrations
func (r *Runner) getAppliedMigrations(ctx context.Context) ([]Migration, error) {
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "applied_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var migrations []Migration
	if err := cursor.All(ctx, &migrations); err != nil {
		return nil, err
	}

	return migrations, nil
}

// getLastMigration returns the last applied migration
func (r *Runner) getLastMigration(ctx context.Context) (*Migration, error) {
	var migration Migration
	opts := options.FindOne().SetSort(bson.D{{Key: "applied_at", Value: -1}})
	err := r.collection.FindOne(ctx, bson.M{}, opts).Decode(&migration)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &migration, nil
}

// recordMigration records a migration as applied
func (r *Runner) recordMigration(ctx context.Context, def MigrationDefinition) error {
	migration := Migration{
		Version:     def.Version,
		Name:        def.Name,
		Description: def.Description,
		AppliedAt:   time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, migration)
	return err
}

// removeMigration removes a migration record
func (r *Runner) removeMigration(ctx context.Context, version string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"version": version})
	return err
}
