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
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM incoming_docs").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count incoming documents: %w", err)
	}

	query := `
		SELECT 
			i.id, i.incoming_no, i.incoming_date, i.received_date, i.status, 
			i.doc_details_id, i.folder_id, i.created_by, i.updated_by, i.approver_id, 
			i.approver_date, i.remark, i.updated_at,
			d.doc_no, d.doc_name,
			v.doc_path,
			u1.firstname || ' ' || u1.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN doc_details d ON i.doc_details_id = d.id
		LEFT JOIN versions v ON i.doc_details_id = v.doc_details_id AND i.folder_id = v.folder_id
		LEFT JOIN users u1 ON i.created_by = u1.id
		LEFT JOIN users u2 ON i.updated_by = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		ORDER BY i.updated_at DESC
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
		var docNo, docName, docPath, creatorName, updaterName, approverName *string
		if err := rows.Scan(
			&doc.ID, &doc.IncomingNo, &doc.IncomingDate, &doc.ReceivedDate, &doc.Status,
			&doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.ApproverID,
			&doc.ApproverDate, &doc.Remark, &doc.UpdatedAt,
			&docNo, &docName, &docPath, &creatorName, &updaterName, &approverName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan incoming document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if creatorName != nil { doc.CreatorName = *creatorName }
		if updaterName != nil { doc.UpdaterName = *updaterName }
		if approverName != nil { doc.ApproverName = *approverName }

		documents = append(documents, doc)
	}

	return documents, total, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDoc, error) {
	query := `
		SELECT 
			i.id, i.incoming_no, i.incoming_date, i.received_date, i.status, 
			i.doc_details_id, i.folder_id, i.created_by, i.updated_by, i.approver_id, 
			i.approver_date, i.remark, i.updated_at,
			d.doc_no, d.doc_name,
			v.doc_path,
			u1.firstname || ' ' || u1.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN doc_details d ON i.doc_details_id = d.id
		LEFT JOIN versions v ON i.doc_details_id = v.doc_details_id AND i.folder_id = v.folder_id
		LEFT JOIN users u1 ON i.created_by = u1.id
		LEFT JOIN users u2 ON i.updated_by = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		WHERE i.id = $1
	`

	var doc domain.IncomingDoc
	var docNo, docName, docPath, creatorName, updaterName, approverName *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.IncomingNo, &doc.IncomingDate, &doc.ReceivedDate, &doc.Status,
		&doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.ApproverID,
		&doc.ApproverDate, &doc.Remark, &doc.UpdatedAt,
		&docNo, &docName, &docPath, &creatorName, &updaterName, &approverName,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("incoming document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming document: %w", err)
	}

	if docNo != nil { doc.DocNo = *docNo }
	if docName != nil { doc.DocName = *docName }
	if docPath != nil { doc.DocPath = *docPath }
	if creatorName != nil { doc.CreatorName = *creatorName }
	if updaterName != nil { doc.UpdaterName = *updaterName }
	if approverName != nil { doc.ApproverName = *approverName }

	return &doc, nil
}

func (r *postgresRepository) FindByReceiverID(ctx context.Context, receiverID uuid.UUID) ([]domain.IncomingDoc, error) {
	query := `
		SELECT 
			i.id, i.incoming_no, i.incoming_date, i.received_date, i.status, 
			i.doc_details_id, i.folder_id, i.created_by, i.updated_by, i.approver_id, 
			i.approver_date, i.remark, i.updated_at,
			d.doc_no, d.doc_name,
			v.doc_path,
			u1.firstname || ' ' || u1.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN doc_details d ON i.doc_details_id = d.id
		LEFT JOIN versions v ON i.doc_details_id = v.doc_details_id AND i.folder_id = v.folder_id
		LEFT JOIN users u1 ON i.created_by = u1.id
		LEFT JOIN users u2 ON i.updated_by = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		WHERE i.updated_by = $1
		ORDER BY i.updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query, receiverID)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by receiver: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		var docNo, docName, docPath, creatorName, updaterName, approverName *string
		if err := rows.Scan(
			&doc.ID, &doc.IncomingNo, &doc.IncomingDate, &doc.ReceivedDate, &doc.Status,
			&doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.ApproverID,
			&doc.ApproverDate, &doc.Remark, &doc.UpdatedAt,
			&docNo, &docName, &docPath, &creatorName, &updaterName, &approverName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if creatorName != nil { doc.CreatorName = *creatorName }
		if updaterName != nil { doc.UpdaterName = *updaterName }
		if approverName != nil { doc.ApproverName = *approverName }

		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int) ([]domain.IncomingDoc, int, error) {
	// The new ERD doesn't have dept_id on incoming_docs.
	// For now, let's just return empty until we resolve how department routing works in the new design.
	return []domain.IncomingDoc{}, 0, nil
}

func (r *postgresRepository) FindByStatus(ctx context.Context, status string) ([]domain.IncomingDoc, error) {
	query := `
		SELECT 
			i.id, i.incoming_no, i.incoming_date, i.received_date, i.status, 
			i.doc_details_id, i.folder_id, i.created_by, i.updated_by, i.approver_id, 
			i.approver_date, i.remark, i.updated_at,
			d.doc_no, d.doc_name,
			v.doc_path,
			u1.firstname || ' ' || u1.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN doc_details d ON i.doc_details_id = d.id
		LEFT JOIN versions v ON i.doc_details_id = v.doc_details_id AND i.folder_id = v.folder_id
		LEFT JOIN users u1 ON i.created_by = u1.id
		LEFT JOIN users u2 ON i.updated_by = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		WHERE i.status = $1
		ORDER BY i.updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by status: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		var docNo, docName, docPath, creatorName, updaterName, approverName *string
		if err := rows.Scan(
			&doc.ID, &doc.IncomingNo, &doc.IncomingDate, &doc.ReceivedDate, &doc.Status,
			&doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.ApproverID,
			&doc.ApproverDate, &doc.Remark, &doc.UpdatedAt,
			&docNo, &docName, &docPath, &creatorName, &updaterName, &approverName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if creatorName != nil { doc.CreatorName = *creatorName }
		if updaterName != nil { doc.UpdaterName = *updaterName }
		if approverName != nil { doc.ApproverName = *approverName }

		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByDocID(ctx context.Context, docDetailsID uuid.UUID) ([]domain.IncomingDoc, error) {
	query := `
		SELECT 
			i.id, i.incoming_no, i.incoming_date, i.received_date, i.status, 
			i.doc_details_id, i.folder_id, i.created_by, i.updated_by, i.approver_id, 
			i.approver_date, i.remark, i.updated_at,
			d.doc_no, d.doc_name,
			v.doc_path,
			u1.firstname || ' ' || u1.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name,
			u3.firstname || ' ' || u3.lastname as approver_name
		FROM incoming_docs i
		LEFT JOIN doc_details d ON i.doc_details_id = d.id
		LEFT JOIN versions v ON i.doc_details_id = v.doc_details_id AND i.folder_id = v.folder_id
		LEFT JOIN users u1 ON i.created_by = u1.id
		LEFT JOIN users u2 ON i.updated_by = u2.id
		LEFT JOIN users u3 ON i.approver_id = u3.id
		WHERE i.doc_details_id = $1
		ORDER BY i.updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query, docDetailsID)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by doc details id: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		var docNo, docName, docPath, creatorName, updaterName, approverName *string
		if err := rows.Scan(
			&doc.ID, &doc.IncomingNo, &doc.IncomingDate, &doc.ReceivedDate, &doc.Status,
			&doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.ApproverID,
			&doc.ApproverDate, &doc.Remark, &doc.UpdatedAt,
			&docNo, &docName, &docPath, &creatorName, &updaterName, &approverName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if creatorName != nil { doc.CreatorName = *creatorName }
		if updaterName != nil { doc.UpdaterName = *updaterName }
		if approverName != nil { doc.ApproverName = *approverName }

		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.IncomingDoc) error {
	query := `
		INSERT INTO incoming_docs (
			incoming_no, incoming_date, received_date, status, 
			doc_details_id, folder_id, created_by, updated_by, 
			approver_id, approver_date, remark, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	doc.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, 
		doc.IncomingNo, doc.IncomingDate, doc.ReceivedDate, doc.Status,
		doc.DocDetailsID, doc.FolderID, doc.CreatedBy, doc.UpdatedBy,
		doc.ApproverID, doc.ApproverDate, doc.Remark, doc.UpdatedAt,
	).Scan(&doc.ID)
	if err != nil {
		return fmt.Errorf("failed to create incoming document: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id uuid.UUID, doc *domain.IncomingDoc) error {
	query := `
		UPDATE incoming_docs 
		SET updated_by = $1, approver_id = $2, received_date = $3, approver_date = $4, remark = $5, status = $6, updated_at = $7
		WHERE id = $8
	`
	doc.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query, 
		doc.UpdatedBy, doc.ApproverID, doc.ReceivedDate, doc.ApproverDate, doc.Remark, doc.Status, doc.UpdatedAt, id)
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
		SET status = $1, updated_at = NOW()
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
