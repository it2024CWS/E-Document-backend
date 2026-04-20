package sector

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

func (r *postgresRepository) FindAll(ctx context.Context, limit, offset int, search string) ([]domain.Sector, error) {
	query := `
		SELECT s.id, s.name, s.dept_id, s.created_at, s.updated_at, COALESCE(d.dept_name, '') as dept_name
		FROM sectors s
		LEFT JOIN departments d ON s.dept_id = d.id
		WHERE s.name ILIKE $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, "%"+search+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find sectors: %w", err)
	}
	defer rows.Close()

	var sectors []domain.Sector
	for rows.Next() {
		var sector domain.Sector
		err := rows.Scan(
			&sector.ID,
			&sector.Name,
			&sector.DeptID,
			&sector.CreatedAt,
			&sector.UpdatedAt,
			&sector.DeptName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sector: %w", err)
		}
		sectors = append(sectors, sector)
	}

	return sectors, rows.Err()
}

func (r *postgresRepository) Count(ctx context.Context, search string) (int, error) {
	query := `SELECT COUNT(*) FROM sectors WHERE name ILIKE $1`
	var count int
	err := r.pool.QueryRow(ctx, query, "%"+search+"%").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sectors: %w", err)
	}
	return count, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*domain.Sector, error) {
	sectorID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid sector ID format: %w", err)
	}

	query := `
		SELECT s.id, s.name, s.dept_id, s.created_at, s.updated_at, COALESCE(d.dept_name, '') as dept_name
		FROM sectors s
		LEFT JOIN departments d ON s.dept_id = d.id
		WHERE s.id = $1
	`

	var sector domain.Sector
	err = r.pool.QueryRow(ctx, query, sectorID).Scan(
		&sector.ID,
		&sector.Name,
		&sector.DeptID,
		&sector.CreatedAt,
		&sector.UpdatedAt,
		&sector.DeptName,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("sector not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find sector: %w", err)
	}

	return &sector, nil
}

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID string) ([]domain.Sector, error) {
	dID, err := uuid.Parse(deptID)
	if err != nil {
		return nil, fmt.Errorf("invalid department ID format: %w", err)
	}

	query := `
		SELECT s.id, s.name, s.dept_id, s.created_at, s.updated_at, COALESCE(d.dept_name, '') as dept_name
		FROM sectors s
		LEFT JOIN departments d ON s.dept_id = d.id
		WHERE s.dept_id = $1
		ORDER BY s.name ASC
	`

	rows, err := r.pool.Query(ctx, query, dID)
	if err != nil {
		return nil, fmt.Errorf("failed to find sectors by department: %w", err)
	}
	defer rows.Close()

	var sectors []domain.Sector
	for rows.Next() {
		var sector domain.Sector
		err := rows.Scan(
			&sector.ID,
			&sector.Name,
			&sector.DeptID,
			&sector.CreatedAt,
			&sector.UpdatedAt,
			&sector.DeptName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sector: %w", err)
		}
		sectors = append(sectors, sector)
	}

	return sectors, rows.Err()
}

func (r *postgresRepository) Create(ctx context.Context, sector *domain.Sector) error {
	query := `
		INSERT INTO sectors (id, name, dept_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	if sector.ID == uuid.Nil {
		sector.ID = uuid.New()
	}
	sector.CreatedAt = time.Now()
	sector.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query,
		sector.ID,
		sector.Name,
		sector.DeptID,
		sector.CreatedAt,
		sector.UpdatedAt,
	).Scan(&sector.ID, &sector.CreatedAt, &sector.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create sector: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id string, sector *domain.Sector) error {
	sectorID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid sector ID format: %w", err)
	}

	query := `
		UPDATE sectors 
		SET name = $1, dept_id = $2, updated_at = $3 
		WHERE id = $4
	`

	sector.UpdatedAt = time.Now()
	result, err := r.pool.Exec(ctx, query, sector.Name, sector.DeptID, sector.UpdatedAt, sectorID)
	if err != nil {
		return fmt.Errorf("failed to update sector: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sector not found")
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	sectorID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid sector ID format: %w", err)
	}

	query := `DELETE FROM sectors WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, sectorID)
	if err != nil {
		return fmt.Errorf("failed to delete sector: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sector not found")
	}

	return nil
}
