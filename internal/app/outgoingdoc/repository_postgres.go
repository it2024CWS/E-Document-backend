package outgoingdoc

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

// NewPostgresRepository creates a new Postgres repository
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) FindAll(ctx context.Context, limit, offset int) ([]domain.OutgoingDoc, int, error) {
	// Get total count first
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM outgoing_docs").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count outgoing documents: %w", err)
	}

	query := `
		SELECT 
			o.id, o.outgoing_no, o.doc_id, o.user_id, o.dept_id, o.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u.firstname || ' ' || u.lastname as user_name
		FROM outgoing_docs o
		LEFT JOIN docs d ON o.doc_id = d.id
		LEFT JOIN users u ON o.user_id = u.id
		ORDER BY o.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find outgoing documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.OutgoingDoc
	for rows.Next() {
		var doc domain.OutgoingDoc
		var docNo, docName, docPath, docType, userName *string
		if err := rows.Scan(
			&doc.ID, &doc.OutgoingNo, &doc.DocID, &doc.UserID, &doc.DepartmentID, &doc.CreatedAt,
			&docNo, &docName, &docPath, &docType, &userName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan outgoing document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if docType != nil { doc.Type = *docType }
		if userName != nil { doc.UserName = *userName }

		documents = append(documents, doc)
	}

	return documents, total, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDoc, error) {
	query := `
		SELECT 
			o.id, o.outgoing_no, o.doc_id, o.user_id, o.dept_id, o.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u.firstname || ' ' || u.lastname as user_name
		FROM outgoing_docs o
		LEFT JOIN docs d ON o.doc_id = d.id
		LEFT JOIN users u ON o.user_id = u.id
		WHERE o.id = $1
	`

	var doc domain.OutgoingDoc
	var docNo, docName, docPath, docType, userName *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.OutgoingNo, &doc.DocID, &doc.UserID, &doc.DepartmentID, &doc.CreatedAt,
		&docNo, &docName, &docPath, &docType, &userName,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("outgoing document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find outgoing document: %w", err)
	}

	if docNo != nil { doc.DocNo = *docNo }
	if docName != nil { doc.DocName = *docName }
	if docPath != nil { doc.DocPath = *docPath }
	if docType != nil { doc.Type = *docType }
	if userName != nil { doc.UserName = *userName }

	return &doc, nil
}

func (r *postgresRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.OutgoingDoc, error) {
	query := `
		SELECT 
			o.id, o.outgoing_no, o.doc_id, o.user_id, o.dept_id, o.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u.firstname || ' ' || u.lastname as user_name
		FROM outgoing_docs o
		LEFT JOIN docs d ON o.doc_id = d.id
		LEFT JOIN users u ON o.user_id = u.id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find outgoing documents by user: %w", err)
	}
	defer rows.Close()

	var documents []domain.OutgoingDoc
	for rows.Next() {
		var doc domain.OutgoingDoc
		var docNo, docName, docPath, docType, userName *string
		if err := rows.Scan(
			&doc.ID, &doc.OutgoingNo, &doc.DocID, &doc.UserID, &doc.DepartmentID, &doc.CreatedAt,
			&docNo, &docName, &docPath, &docType, &userName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outgoing document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if docType != nil { doc.Type = *docType }
		if userName != nil { doc.UserName = *userName }

		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int) ([]domain.OutgoingDoc, int, error) {
	// Get total count first
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM outgoing_docs WHERE dept_id = $1", deptID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count outgoing documents by department: %w", err)
	}

	query := `
		SELECT 
			o.id, o.outgoing_no, o.doc_id, o.user_id, o.dept_id, o.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u.firstname || ' ' || u.lastname as user_name
		FROM outgoing_docs o
		LEFT JOIN docs d ON o.doc_id = d.id
		LEFT JOIN users u ON o.user_id = u.id
		WHERE o.dept_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, deptID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find outgoing documents by department: %w", err)
	}
	defer rows.Close()

	var documents []domain.OutgoingDoc
	for rows.Next() {
		var doc domain.OutgoingDoc
		var docNo, docName, docPath, docType, userName *string
		if err := rows.Scan(
			&doc.ID, &doc.OutgoingNo, &doc.DocID, &doc.UserID, &doc.DepartmentID, &doc.CreatedAt,
			&docNo, &docName, &docPath, &docType, &userName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan outgoing document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if docType != nil { doc.Type = *docType }
		if userName != nil { doc.UserName = *userName }

		documents = append(documents, doc)
	}

	return documents, total, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.OutgoingDoc) error {
	query := `
		INSERT INTO outgoing_docs (outgoing_no, doc_id, user_id, dept_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	doc.CreatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, doc.OutgoingNo, doc.DocID, doc.UserID, doc.DepartmentID, doc.CreatedAt).
		Scan(&doc.ID, &doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create outgoing document: %w", err)
	}

	return nil
}
