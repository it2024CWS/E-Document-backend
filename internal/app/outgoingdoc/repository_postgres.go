package outgoingdoc

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

// NewPostgresRepository creates a new Postgres repository
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) FindAll(ctx context.Context) ([]domain.OutgoingDoc, error) {
	query := `
		SELECT id, outgoing_no, doc_id, user_id, created_at
		FROM outgoing_docs
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find outgoing documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.OutgoingDoc
	for rows.Next() {
		var doc domain.OutgoingDoc
		if err := rows.Scan(&doc.ID, &doc.OutgoingNo, &doc.DocID, &doc.UserID, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan outgoing document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id int) (*domain.OutgoingDoc, error) {
	query := `
		SELECT id, outgoing_no, doc_id, user_id, created_at
		FROM outgoing_docs
		WHERE id = $1
	`

	var doc domain.OutgoingDoc
	err := r.pool.QueryRow(ctx, query, id).Scan(&doc.ID, &doc.OutgoingNo, &doc.DocID, &doc.UserID, &doc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("outgoing document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find outgoing document: %w", err)
	}

	return &doc, nil
}

func (r *postgresRepository) FindByUserID(ctx context.Context, userID string) ([]domain.OutgoingDoc, error) {
	query := `
		SELECT id, outgoing_no, doc_id, user_id, created_at
		FROM outgoing_docs
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find outgoing documents by user: %w", err)
	}
	defer rows.Close()

	var documents []domain.OutgoingDoc
	for rows.Next() {
		var doc domain.OutgoingDoc
		if err := rows.Scan(&doc.ID, &doc.OutgoingNo, &doc.DocID, &doc.UserID, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan outgoing document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.OutgoingDoc) error {
	query := `
		INSERT INTO outgoing_docs (outgoing_no, doc_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	doc.CreatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, doc.OutgoingNo, doc.DocID, doc.UserID, doc.CreatedAt).
		Scan(&doc.ID, &doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create outgoing document: %w", err)
	}

	return nil
}
