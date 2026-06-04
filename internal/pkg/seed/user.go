package seed

import (
	"context"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// adminUserDef defines a user to be seeded.
type adminUserDef struct {
	username  string
	email     string
	password  string
	firstname string
	lastname  string
	phone     string
	roleName  string
}

// SeedUsers seeds the system admin user (and any other default users).
// It is idempotent: existing users (matched by email) are skipped.
func SeedUsers(ctx context.Context, userRepo user.Repository, pool *pgxpool.Pool) error {
	users := []adminUserDef{
		{
			username:  "admin",
			email:     "admin@edocument.com",
			password:  "123456",
			firstname: "System",
			lastname:  "Admin",
			phone:     "000-000-0000",
			roleName:  "Admin",
		},
	}

	for _, u := range users {
		if err := seedOneUser(ctx, userRepo, pool, u); err != nil {
			return err
		}
	}

	return nil
}

// seedOneUser inserts a single user if it does not already exist.
func seedOneUser(ctx context.Context, userRepo user.Repository, pool *pgxpool.Pool, def adminUserDef) error {
	// Skip if the user already exists
	existing, err := userRepo.FindByEmail(ctx, def.email)
	if err == nil && existing != nil {
		logger.Infof("⚠ User '%s' already exists, skipping", def.username)
		return nil
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(def.password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Resolve role ID
	var roleID uuid.UUID
	if err := pool.QueryRow(ctx,
		`SELECT id FROM user_roles WHERE role_name = $1`, def.roleName,
	).Scan(&roleID); err != nil {
		return err
	}

	isActive := true
	newUser := &domain.User{
		Username:  def.username,
		Email:     def.email,
		Password:  string(hashed),
		Firstname: def.firstname,
		Lastname:  def.lastname,
		Phone:     def.phone,
		RoleID:    &roleID,
		IsActive:  isActive,
	}

	if err := userRepo.Create(ctx, newUser); err != nil {
		return err
	}

	logger.Infof("✓ User '%s' (%s) created with role '%s'", def.username, def.email, def.roleName)
	return nil
}
