package migrations

import (
	"e-document-backend/internal/migration"
)

// GetAll returns all registered migrations
func GetAll() []migration.MigrationDefinition {
	return []migration.MigrationDefinition{
		// Register your migrations here in order
		// Comment out migrations you don't want to run:
		// Migration001_AddPhoneFieldToUsers(),
		// Migration002_CreateEmailIndex(),
	}
}

// RegisterMigrations registers all migrations with the runner
func RegisterMigrations(runner *migration.Runner) {
	for _, m := range GetAll() {
		runner.Register(m)
	}
}
