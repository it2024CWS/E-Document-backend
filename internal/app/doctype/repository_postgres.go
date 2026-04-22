package doctype

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

func (r *postgresRepository) FindAll(ctx context.Context) ([]domain.DocType, error) {
	query := `SELECT id, type_name, created_at FROM doc_types ORDER BY id ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find doc types: %w", err)
	}
	defer rows.Close()

	var docTypes []domain.DocType
	for rows.Next() {
		var docType domain.DocType
		if err := rows.Scan(&docType.ID, &docType.TypeName, &docType.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan doc type: %w", err)
		}
		docTypes = append(docTypes, docType)
	}

	return docTypes, rows.Err()
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*domain.DocType, error) {
	query := `SELECT id, type_name, created_at FROM doc_types WHERE id = $1`

	var docType domain.DocType
	err := r.pool.QueryRow(ctx, query, id).Scan(&docType.ID, &docType.TypeName, &docType.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("doc type not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find doc type: %w", err)
	}

	return &docType, nil
}

func (r *postgresRepository) Create(ctx context.Context, docType *domain.DocType) error {
	query := `
		INSERT INTO doc_types (type_name, created_at)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	docType.CreatedAt = time.Now()
	err := r.pool.QueryRow(ctx, query, docType.TypeName, docType.CreatedAt).Scan(&docType.ID, &docType.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create doc type: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id string, docType *domain.DocType) error {
	query := `UPDATE doc_types SET type_name = $1 WHERE id = $2`

	result, err := r.pool.Exec(ctx, query, docType.TypeName, id)
	if err != nil {
		return fmt.Errorf("failed to update doc type: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("doc type not found")
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM doc_types WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete doc type: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("doc type not found")
	}

	return nil
}
