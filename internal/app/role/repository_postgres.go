package role

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL role repository
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{
		pool: pool,
	}
}

// FindAll retrieves all user roles with pagination and search
func (r *postgresRepository) FindAll(ctx context.Context, limit, offset int, search string) ([]domain.UserRole, error) {
	query := `
		SELECT id, role_name, created_at
		FROM user_roles
		WHERE role_name ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, "%"+search+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find roles: %w", err)
	}
	defer rows.Close()

	var roles []domain.UserRole
	for rows.Next() {
		var role domain.UserRole
		err := rows.Scan(
			&role.ID,
			&role.RoleName,
			&role.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return roles, nil
}

// Count returns the total number of roles matching the search criteria
func (r *postgresRepository) Count(ctx context.Context, search string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_roles
		WHERE role_name ILIKE $1
	`

	var count int
	err := r.pool.QueryRow(ctx, query, "%"+search+"%").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count roles: %w", err)
	}

	return count, nil
}

// FindByID retrieves a role by ID
func (r *postgresRepository) FindByID(ctx context.Context, id string) (*domain.UserRole, error) {
	query := `
		SELECT id, role_name, created_at
		FROM user_roles
		WHERE id = $1
	`

	var role domain.UserRole
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&role.ID,
		&role.RoleName,
		&role.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to find role: %w", err)
	}

	return &role, nil
}

// Create inserts a new role
func (r *postgresRepository) Create(ctx context.Context, role *domain.UserRole) error {
	query := `
		INSERT INTO user_roles (role_name, created_at)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	role.CreatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query,
		role.RoleName,
		role.CreatedAt,
	).Scan(&role.ID, &role.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// Update updates a role by ID
func (r *postgresRepository) Update(ctx context.Context, id string, role *domain.UserRole) error {
	query := `
		UPDATE user_roles
		SET role_name = $1
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, role.RoleName, id)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// Delete deletes a role by ID
func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM user_roles WHERE id = $1"

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}
