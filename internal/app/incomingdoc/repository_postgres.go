package incomingdoc

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

func (r *postgresRepository) FindAll(ctx context.Context) ([]domain.IncomingDoc, error) {
	query := `
		SELECT id, incoming_no, doc_id, sender_id, receiver_id, approver_id, received_date, approver_date, remark, status, created_at
		FROM incoming_docs
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := rows.Scan(&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID, &doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id int) (*domain.IncomingDoc, error) {
	query := `
		SELECT id, incoming_no, doc_id, sender_id, receiver_id, approver_id, received_date, approver_date, remark, status, created_at
		FROM incoming_docs
		WHERE id = $1
	`

	var doc domain.IncomingDoc
	err := r.pool.QueryRow(ctx, query, id).Scan(&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID, &doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("incoming document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming document: %w", err)
	}

	return &doc, nil
}

func (r *postgresRepository) FindByReceiverID(ctx context.Context, receiverID string) ([]domain.IncomingDoc, error) {
	query := `
		SELECT id, incoming_no, doc_id, sender_id, receiver_id, approver_id, received_date, approver_date, remark, status, created_at
		FROM incoming_docs
		WHERE receiver_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, receiverID)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by receiver: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := rows.Scan(&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID, &doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByStatus(ctx context.Context, status string) ([]domain.IncomingDoc, error) {
	query := `
		SELECT id, incoming_no, doc_id, sender_id, receiver_id, approver_id, received_date, approver_date, remark, status, created_at
		FROM incoming_docs
		WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by status: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := rows.Scan(&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID, &doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.IncomingDoc) error {
	query := `
		INSERT INTO incoming_docs (incoming_no, doc_id, sender_id, receiver_id, approver_id, received_date, approver_date, remark, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	doc.CreatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, doc.IncomingNo, doc.DocID, doc.SenderID, doc.ReceiverID, doc.ApproverID, doc.ReceivedDate, doc.ApproverDate, doc.Remark, doc.Status, doc.CreatedAt).
		Scan(&doc.ID, &doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create incoming document: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id int, doc *domain.IncomingDoc) error {
	query := `
		UPDATE incoming_docs 
		SET receiver_id = $1, approver_id = $2, received_date = $3, approver_date = $4, remark = $5, status = $6
		WHERE id = $7
	`

	result, err := r.pool.Exec(ctx, query, doc.ReceiverID, doc.ApproverID, doc.ReceivedDate, doc.ApproverDate, doc.Remark, doc.Status, id)
	if err != nil {
		return fmt.Errorf("failed to update incoming document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("incoming document not found")
	}

	return nil
}

func (r *postgresRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE incoming_docs 
		SET status = $1
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update incoming document status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("incoming document not found")
	}

	return nil
}
