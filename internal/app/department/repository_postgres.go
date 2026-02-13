package department

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

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) FindAll(ctx context.Context) ([]domain.Department, error) {
	query := `SELECT id, dept_name, created_at FROM departments ORDER BY id ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find departments: %w", err)
	}
	defer rows.Close()

	var departments []domain.Department
	for rows.Next() {
		var dept domain.Department
		if err := rows.Scan(&dept.ID, &dept.DeptName, &dept.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan department: %w", err)
		}
		departments = append(departments, dept)
	}

	return departments, rows.Err()
}

func (r *postgresRepository) FindByID(ctx context.Context, id int) (*domain.Department, error) {
	query := `SELECT id, dept_name, created_at FROM departments WHERE id = $1`

	var dept domain.Department
	err := r.pool.QueryRow(ctx, query, id).Scan(&dept.ID, &dept.DeptName, &dept.CreatedAt)
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
		INSERT INTO departments (dept_name, created_at)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	dept.CreatedAt = time.Now()
	err := r.pool.QueryRow(ctx, query, dept.DeptName, dept.CreatedAt).Scan(&dept.ID, &dept.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create department: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id int, dept *domain.Department) error {
	query := `UPDATE departments SET dept_name = $1 WHERE id = $2`

	result, err := r.pool.Exec(ctx, query, dept.DeptName, id)
	if err != nil {
		return fmt.Errorf("failed to update department: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("department not found")
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM departments WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("department not found")
	}

	return nil
}
