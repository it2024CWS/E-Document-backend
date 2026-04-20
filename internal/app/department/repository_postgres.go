package department

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) FindAll(ctx context.Context, limit, offset int, search string) ([]domain.Department, error) {
	query := `
		SELECT id, dept_name, COALESCE(description, '') as description, created_at, updated_at
		FROM departments
		WHERE dept_name ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, "%"+search+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find departments: %w", err)
	}
	defer rows.Close()

	var departments []domain.Department
	for rows.Next() {
		var dept domain.Department
		err := rows.Scan(
			&dept.ID,
			&dept.DeptName,
			&dept.Description,
			&dept.CreatedAt,
			&dept.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan department: %w", err)
		}
		departments = append(departments, dept)
	}

	return departments, rows.Err()
}

func (r *postgresRepository) Count(ctx context.Context, search string) (int, error) {
	query := `SELECT COUNT(*) FROM departments WHERE dept_name ILIKE $1`
	var count int
	err := r.pool.QueryRow(ctx, query, "%"+search+"%").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count departments: %w", err)
	}
	return count, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*domain.Department, error) {
	deptID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid department ID format: %w", err)
	}

	query := `
		SELECT id, dept_name, COALESCE(description, '') as description, created_at, updated_at 
		FROM departments 
		WHERE id = $1
	`

	var dept domain.Department
	err = r.pool.QueryRow(ctx, query, deptID).Scan(
		&dept.ID,
		&dept.DeptName,
		&dept.Description,
		&dept.CreatedAt,
		&dept.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("department not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find department: %w", err)
	}

	return &dept, nil
}

func (r *postgresRepository) Create(ctx context.Context, dept *domain.Department) error {
	query := `
		INSERT INTO departments (id, dept_name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	if dept.ID == uuid.Nil {
		dept.ID = uuid.New()
	}
	dept.CreatedAt = time.Now()
	dept.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query,
		dept.ID,
		dept.DeptName,
		dept.Description,
		dept.CreatedAt,
		dept.UpdatedAt,
	).Scan(&dept.ID, &dept.CreatedAt, &dept.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create department: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id string, dept *domain.Department) error {
	deptID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid department ID format: %w", err)
	}

	query := `
		UPDATE departments 
		SET dept_name = $1, description = $2, updated_at = $3 
		WHERE id = $4
	`

	dept.UpdatedAt = time.Now()
	result, err := r.pool.Exec(ctx, query, dept.DeptName, dept.Description, dept.UpdatedAt, deptID)
	if err != nil {
		return fmt.Errorf("failed to update department: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("department not found")
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	deptID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid department ID format: %w", err)
	}

	query := `DELETE FROM departments WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, deptID)
	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("department not found")
	}

	return nil
}
