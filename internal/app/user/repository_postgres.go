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
			id, username, email, phone, firstname, lastname, nickname,
			password, role_id, department_id, sector_id, is_active, profile_picture,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
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
		user.Firstname,
		user.Lastname,
		user.Nickname,
		user.Password,
		user.RoleID,
		user.DepartmentID,
		user.SectorID,
		user.IsActive,
		user.ProfilePicture,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by ID with joined role, department, and sector names
func (r *postgresRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT 
			u.id, u.username, u.email, u.phone, u.firstname, u.lastname, u.nickname,
			u.password, u.role_id, u.department_id, u.sector_id, u.is_active, u.profile_picture,
			u.created_at, u.updated_at,
			COALESCE(ur.role_name, '') as role_name, 
			COALESCE(d.dept_name, '') as dept_name, 
			COALESCE(s.name, '') as sector_name
		FROM users u
		LEFT JOIN user_roles ur ON u.role_id = ur.id
		LEFT JOIN departments d ON u.department_id = d.id
		LEFT JOIN sectors s ON u.sector_id = s.id
		WHERE u.id = $1
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
		&user.Firstname,
		&user.Lastname,
		&user.Nickname,
		&user.Password,
		&user.RoleID,
		&user.DepartmentID,
		&user.SectorID,
		&user.IsActive,
		&user.ProfilePicture,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.RoleName,
		&user.DepartmentName,
		&user.SectorName,
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
		SELECT 
			u.id, u.username, u.email, u.phone, u.firstname, u.lastname, u.nickname,
			u.password, u.role_id, u.department_id, u.sector_id, u.is_active, u.profile_picture,
			u.created_at, u.updated_at
		FROM users u
		WHERE u.username = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.Firstname,
		&user.Lastname,
		&user.Nickname,
		&user.Password,
		&user.RoleID,
		&user.DepartmentID,
		&user.SectorID,
		&user.IsActive,
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
		SELECT 
			u.id, u.username, u.email, u.phone, u.firstname, u.lastname, u.nickname,
			u.password, u.role_id, u.department_id, u.sector_id, u.is_active, u.profile_picture,
			u.created_at, u.updated_at
		FROM users u
		WHERE u.email = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.Firstname,
		&user.Lastname,
		&user.Nickname,
		&user.Password,
		&user.RoleID,
		&user.DepartmentID,
		&user.SectorID,
		&user.IsActive,
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

// FindAll retrieves all users with pagination and search
func (r *postgresRepository) FindAll(ctx context.Context, skip int, limit int, search string, currentUserID string) ([]domain.User, error) {
	query := `
		SELECT 
			u.id, u.username, u.email, u.phone, u.firstname, u.lastname, u.nickname,
			u.password, u.role_id, u.department_id, u.sector_id, u.is_active, u.profile_picture,
			u.created_at, u.updated_at,
			COALESCE(ur.role_name, '') as role_name, 
			COALESCE(d.dept_name, '') as dept_name, 
			COALESCE(s.name, '') as sector_name
		FROM users u
		LEFT JOIN user_roles ur ON u.role_id = ur.id
		LEFT JOIN departments d ON u.department_id = d.id
		LEFT JOIN sectors s ON u.sector_id = s.id
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	argCount := 1

	// Add search filter
	if search != "" {
		query += fmt.Sprintf(" AND (u.username ILIKE $%d OR u.email ILIKE $%d OR u.firstname ILIKE $%d OR u.lastname ILIKE $%d)", argCount, argCount, argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	// Add ordering (show current user first if provided)
	if currentUserID != "" {
		userID, err := uuid.Parse(currentUserID)
		if err == nil {
			query += fmt.Sprintf(" ORDER BY (u.id = $%d) DESC, u.created_at DESC", argCount)
			args = append(args, userID)
			argCount++
		} else {
			query += " ORDER BY u.created_at DESC"
		}
	} else {
		query += " ORDER BY u.created_at DESC"
	}

	// Add pagination
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
			&user.Firstname,
			&user.Lastname,
			&user.Nickname,
			&user.Password,
			&user.RoleID,
			&user.DepartmentID,
			&user.SectorID,
			&user.IsActive,
			&user.ProfilePicture,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.RoleName,
			&user.DepartmentName,
			&user.SectorName,
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

// Count returns the total number of users
func (r *postgresRepository) Count(ctx context.Context, search string, currentUserID string) (int, error) {
	query := "SELECT COUNT(*) FROM users WHERE 1=1"

	args := make([]interface{}, 0)
	argCount := 1

	// Add search filter
	if search != "" {
		query += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d OR firstname ILIKE $%d OR lastname ILIKE $%d)", argCount, argCount, argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
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
		    firstname = $4,
		    lastname = $5,
		    nickname = $6,
		    password = $7,
		    role_id = $8,
		    department_id = $9,
		    sector_id = $10,
		    is_active = $11,
		    profile_picture = $12,
		    updated_at = $13
		WHERE id = $14
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
		user.Firstname,
		user.Lastname,
		user.Nickname,
		user.Password,
		user.RoleID,
		user.DepartmentID,
		user.SectorID,
		user.IsActive,
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
