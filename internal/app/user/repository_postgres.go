package user

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// postgresRepository implements the Repository interface for PostgreSQL
type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL user repository
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{
		pool: pool,
	}
}

// Create inserts a new user into the database
func (r *postgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, username, email, phone, first_name, last_name,
			password, role, department_id, sector_id, profile_picture,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		RETURNING id, created_at, updated_at
	`

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Phone,
		user.FirstName,
		user.LastName,
		user.Password,
		user.Role,
		user.DepartmentID,
		user.SectorID,
		user.ProfilePicture,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by ID
func (r *postgresRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, phone, first_name, last_name,
		       password, role, department_id, sector_id, profile_picture,
		       created_at, updated_at
		FROM users
		WHERE id = $1
	`

	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var user domain.User
	err = r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.Role,
		&user.DepartmentID,
		&user.SectorID,
		&user.ProfilePicture,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindByUsername retrieves a user by username
func (r *postgresRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, phone, first_name, last_name,
		       password, role, department_id, sector_id, profile_picture,
		       created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.Role,
		&user.DepartmentID,
		&user.SectorID,
		&user.ProfilePicture,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindByEmail retrieves a user by email
func (r *postgresRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, phone, first_name, last_name,
		       password, role, department_id, sector_id, profile_picture,
		       created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.Role,
		&user.DepartmentID,
		&user.SectorID,
		&user.ProfilePicture,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// FindAll retrieves all users with pagination and search (excluding current user)
func (r *postgresRepository) FindAll(ctx context.Context, skip int, limit int, search string, currentUserID string) ([]domain.User, error) {
	query := `
		SELECT id, username, email, phone, first_name, last_name,
		       password, role, department_id, sector_id, profile_picture,
		       created_at, updated_at
		FROM users
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	argCount := 1

	// Add search filter
	if search != "" {
		query += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	// Exclude current user
	if currentUserID != "" {
		userID, err := uuid.Parse(currentUserID)
		if err == nil {
			query += fmt.Sprintf(" AND id != $%d", argCount)
			args = append(args, userID)
			argCount++
		}
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, skip)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Phone,
			&user.FirstName,
			&user.LastName,
			&user.Password,
			&user.Role,
			&user.DepartmentID,
			&user.SectorID,
			&user.ProfilePicture,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

// Count returns the total number of users (excluding current user)
func (r *postgresRepository) Count(ctx context.Context, search string, currentUserID string) (int, error) {
	query := "SELECT COUNT(*) FROM users WHERE 1=1"

	args := make([]interface{}, 0)
	argCount := 1

	// Add search filter
	if search != "" {
		query += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	// Exclude current user
	if currentUserID != "" {
		userID, err := uuid.Parse(currentUserID)
		if err == nil {
			query += fmt.Sprintf(" AND id != $%d", argCount)
			args = append(args, userID)
		}
	}

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// Update updates a user by ID
func (r *postgresRepository) Update(ctx context.Context, id string, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $1,
		    email = $2,
		    phone = $3,
		    first_name = $4,
		    last_name = $5,
		    password = $6,
		    role = $7,
		    department_id = $8,
		    sector_id = $9,
		    profile_picture = $10,
		    updated_at = $11
		WHERE id = $12
	`

	userID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	user.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query,
		user.Username,
		user.Email,
		user.Phone,
		user.FirstName,
		user.LastName,
		user.Password,
		user.Role,
		user.DepartmentID,
		user.SectorID,
		user.ProfilePicture,
		user.UpdatedAt,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete deletes a user by ID
func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM users WHERE id = $1"

	userID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	result, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
