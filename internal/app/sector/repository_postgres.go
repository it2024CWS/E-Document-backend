package sector

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

func (r *postgresRepository) FindAll(ctx context.Context) ([]domain.Sector, error) {
	query := `
		SELECT s.id, s.name, s.dept_id, s.created_at
		FROM sectors s
		ORDER BY s.id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find sectors: %w", err)
	}
	defer rows.Close()

	var sectors []domain.Sector
	for rows.Next() {
		var sector domain.Sector
		if err := rows.Scan(&sector.ID, &sector.Name, &sector.DeptID, &sector.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan sector: %w", err)
		}
		sectors = append(sectors, sector)
	}

	return sectors, rows.Err()
}

func (r *postgresRepository) FindByID(ctx context.Context, id int) (*domain.Sector, error) {
	query := `
		SELECT id, name, dept_id, created_at
		FROM sectors
		WHERE id = $1
	`

	var sector domain.Sector
	err := r.pool.QueryRow(ctx, query, id).Scan(&sector.ID, &sector.Name, &sector.DeptID, &sector.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("sector not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find sector: %w", err)
	}

	return &sector, nil
}

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID int) ([]domain.Sector, error) {
	query := `
		SELECT id, name, dept_id, created_at
		FROM sectors
		WHERE dept_id = $1
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query, deptID)
	if err != nil {
		return nil, fmt.Errorf("failed to find sectors by department: %w", err)
	}
	defer rows.Close()

	var sectors []domain.Sector
	for rows.Next() {
		var sector domain.Sector
		if err := rows.Scan(&sector.ID, &sector.Name, &sector.DeptID, &sector.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan sector: %w", err)
		}
		sectors = append(sectors, sector)
	}

	return sectors, rows.Err()
}

func (r *postgresRepository) Create(ctx context.Context, sector *domain.Sector) error {
	query := `
		INSERT INTO sectors (name, dept_id, created_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	sector.CreatedAt = time.Now()
	err := r.pool.QueryRow(ctx, query, sector.Name, sector.DeptID, sector.CreatedAt).
		Scan(&sector.ID, &sector.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create sector: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id int, sector *domain.Sector) error {
	query := `UPDATE sectors SET name = $1, dept_id = $2 WHERE id = $3`

	result, err := r.pool.Exec(ctx, query, sector.Name, sector.DeptID, id)
	if err != nil {
		return fmt.Errorf("failed to update sector: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sector not found")
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM sectors WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sector: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sector not found")
	}

	return nil
}
