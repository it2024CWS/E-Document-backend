package incomingdoc

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

func (r *postgresRepository) FindAll(ctx context.Context, limit, offset int) ([]domain.IncomingDoc, int, error) {
	// Get total count first
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM incoming_docs").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count incoming documents: %w", err)
	}

	query := `
		SELECT 
			i.id, i.incoming_no, i.doc_id, i.sender_id, i.receiver_id, i.approver_id, 
			i.received_date, i.approver_date, i.remark, i.status, i.dept_id, i.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u1.firstname || ' ' || u1.lastname as sender_name,
			u2.firstname || ' ' || u2.lastname as receiver_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN docs d ON i.doc_id = d.id
		LEFT JOIN users u1 ON i.sender_id = u1.id
		LEFT JOIN users u2 ON i.receiver_id = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		ORDER BY i.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find incoming documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		var docNo, docName, docPath, docType, senderName, receiverName, approverName *string
		if err := rows.Scan(
			&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID,
			&doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.DepartmentID, &doc.CreatedAt,
			&docNo, &docName, &docPath, &docType, &senderName, &receiverName, &approverName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan incoming document: %w", err)
		}

		// Map extra fields if they exist
		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if docType != nil { doc.Type = *docType }
		if senderName != nil { doc.SenderName = *senderName }
		if receiverName != nil { doc.ReceiverName = *receiverName }
		if approverName != nil { doc.ApproverName = *approverName }

		documents = append(documents, doc)
	}

	return documents, total, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDoc, error) {
	query := `
		SELECT 
			i.id, i.incoming_no, i.doc_id, i.sender_id, i.receiver_id, i.approver_id, 
			i.received_date, i.approver_date, i.remark, i.status, i.dept_id, i.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u1.firstname || ' ' || u1.lastname as sender_name,
			u2.firstname || ' ' || u2.lastname as receiver_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN docs d ON i.doc_id = d.id
		LEFT JOIN users u1 ON i.sender_id = u1.id
		LEFT JOIN users u2 ON i.receiver_id = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		WHERE i.id = $1
	`

	var doc domain.IncomingDoc
	var docNo, docName, docPath, docType, senderName, receiverName, approverName *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID,
		&doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.DepartmentID, &doc.CreatedAt,
		&docNo, &docName, &docPath, &docType, &senderName, &receiverName, &approverName,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("incoming document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming document: %w", err)
	}

	// Map extra fields
	if docNo != nil { doc.DocNo = *docNo }
	if docName != nil { doc.DocName = *docName }
	if docPath != nil { doc.DocPath = *docPath }
	if docType != nil { doc.Type = *docType }
	if senderName != nil { doc.SenderName = *senderName }
	if receiverName != nil { doc.ReceiverName = *receiverName }
	if approverName != nil { doc.ApproverName = *approverName }

	return &doc, nil
}

func (r *postgresRepository) FindByReceiverID(ctx context.Context, receiverID uuid.UUID) ([]domain.IncomingDoc, error) {
	query := `
		SELECT 
			i.id, i.incoming_no, i.doc_id, i.sender_id, i.receiver_id, i.approver_id, 
			i.received_date, i.approver_date, i.remark, i.status, i.dept_id, i.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u1.firstname || ' ' || u1.lastname as sender_name,
			u2.firstname || ' ' || u2.lastname as receiver_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN docs d ON i.doc_id = d.id
		LEFT JOIN users u1 ON i.sender_id = u1.id
		LEFT JOIN users u2 ON i.receiver_id = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		WHERE i.receiver_id = $1
		ORDER BY i.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, receiverID)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by receiver: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		var docNo, docName, docPath, docType, senderName, receiverName, approverName *string
		if err := rows.Scan(
			&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID,
			&doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.DepartmentID, &doc.CreatedAt,
			&docNo, &docName, &docPath, &docType, &senderName, &receiverName, &approverName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if docType != nil { doc.Type = *docType }
		if senderName != nil { doc.SenderName = *senderName }
		if receiverName != nil { doc.ReceiverName = *receiverName }
		if approverName != nil { doc.ApproverName = *approverName }

		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int) ([]domain.IncomingDoc, int, error) {
	// Get total count first
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM incoming_docs WHERE dept_id = $1", deptID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count incoming documents by department: %w", err)
	}

	query := `
		SELECT 
			i.id, i.incoming_no, i.doc_id, i.sender_id, i.receiver_id, i.approver_id, 
			i.received_date, i.approver_date, i.remark, i.status, i.dept_id, i.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u1.firstname || ' ' || u1.lastname as sender_name,
			u2.firstname || ' ' || u2.lastname as receiver_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN docs d ON i.doc_id = d.id
		LEFT JOIN users u1 ON i.sender_id = u1.id
		LEFT JOIN users u2 ON i.receiver_id = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		WHERE i.dept_id = $1
		ORDER BY i.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, deptID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find incoming documents by department: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		var docNo, docName, docPath, docType, senderName, receiverName, approverName *string
		if err := rows.Scan(
			&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID,
			&doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.DepartmentID, &doc.CreatedAt,
			&docNo, &docName, &docPath, &docType, &senderName, &receiverName, &approverName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan incoming document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if docType != nil { doc.Type = *docType }
		if senderName != nil { doc.SenderName = *senderName }
		if receiverName != nil { doc.ReceiverName = *receiverName }
		if approverName != nil { doc.ApproverName = *approverName }

		documents = append(documents, doc)
	}

	return documents, total, nil
}

func (r *postgresRepository) FindByStatus(ctx context.Context, status string) ([]domain.IncomingDoc, error) {
	query := `
		SELECT 
			i.id, i.incoming_no, i.doc_id, i.sender_id, i.receiver_id, i.approver_id, 
			i.received_date, i.approver_date, i.remark, i.status, i.dept_id, i.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u1.firstname || ' ' || u1.lastname as sender_name,
			u2.firstname || ' ' || u2.lastname as receiver_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN docs d ON i.doc_id = d.id
		LEFT JOIN users u1 ON i.sender_id = u1.id
		LEFT JOIN users u2 ON i.receiver_id = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		WHERE i.status = $1
		ORDER BY i.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by status: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		var docNo, docName, docPath, docType, senderName, receiverName, approverName *string
		if err := rows.Scan(
			&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID,
			&doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.DepartmentID, &doc.CreatedAt,
			&docNo, &docName, &docPath, &docType, &senderName, &receiverName, &approverName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if docType != nil { doc.Type = *docType }
		if senderName != nil { doc.SenderName = *senderName }
		if receiverName != nil { doc.ReceiverName = *receiverName }
		if approverName != nil { doc.ApproverName = *approverName }

		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByDocID(ctx context.Context, docID uuid.UUID) ([]domain.IncomingDoc, error) {
	query := `
		SELECT 
			i.id, i.incoming_no, i.doc_id, i.sender_id, i.receiver_id, i.approver_id, 
			i.received_date, i.approver_date, i.remark, i.status, i.dept_id, i.created_at,
			d.doc_no, d.doc_name, d.doc_path, d.type,
			u1.firstname || ' ' || u1.lastname as sender_name,
			u2.firstname || ' ' || u2.lastname as receiver_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN docs d ON i.doc_id = d.id
		LEFT JOIN users u1 ON i.sender_id = u1.id
		LEFT JOIN users u2 ON i.receiver_id = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		WHERE i.doc_id = $1
		ORDER BY i.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by doc id: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		var docNo, docName, docPath, docType, senderName, receiverName, approverName *string
		if err := rows.Scan(
			&doc.ID, &doc.IncomingNo, &doc.DocID, &doc.SenderID, &doc.ReceiverID, &doc.ApproverID,
			&doc.ReceivedDate, &doc.ApproverDate, &doc.Remark, &doc.Status, &doc.DepartmentID, &doc.CreatedAt,
			&docNo, &docName, &docPath, &docType, &senderName, &receiverName, &approverName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if docType != nil { doc.Type = *docType }
		if senderName != nil { doc.SenderName = *senderName }
		if receiverName != nil { doc.ReceiverName = *receiverName }
		if approverName != nil { doc.ApproverName = *approverName }

		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.IncomingDoc) error {
	query := `
		INSERT INTO incoming_docs (incoming_no, doc_id, sender_id, receiver_id, approver_id, received_date, approver_date, remark, status, dept_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`

	doc.CreatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, doc.IncomingNo, doc.DocID, doc.SenderID, doc.ReceiverID, doc.ApproverID, doc.ReceivedDate, doc.ApproverDate, doc.Remark, doc.Status, doc.DepartmentID, doc.CreatedAt).
		Scan(&doc.ID, &doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create incoming document: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id uuid.UUID, doc *domain.IncomingDoc) error {
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

func (r *postgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
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
