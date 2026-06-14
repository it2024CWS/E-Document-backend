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

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) FindAll(ctx context.Context, limit, offset int) ([]domain.OutgoingDoc, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM outgoing_docs").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count outgoing documents: %w", err)
	}

	query := `
		SELECT
			o.id, o.outgoing_no, o.doc_details_id, o.folder_id, o.created_by, o.updated_by, o.created_at,
			d.doc_no, d.doc_name, v.doc_path, v.file_type,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN users u ON o.created_by = u.id
		LEFT JOIN users u2 ON o.updated_by = u2.id
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
		var docNo, docName, docPath, fileType, creatorName, updaterName *string
		if err := rows.Scan(
			&doc.ID, &doc.OutgoingNo, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt,
			&docNo, &docName, &docPath, &fileType, &creatorName, &updaterName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan outgoing document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if fileType != nil { doc.FileType = *fileType }
		if creatorName != nil { doc.CreatorName = *creatorName }
		if updaterName != nil { doc.UpdaterName = *updaterName }

		documents = append(documents, doc)
	}

	return documents, total, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDoc, error) {
	query := `
		SELECT
			o.id, o.outgoing_no, o.doc_details_id, o.folder_id, o.created_by, o.updated_by, o.created_at,
			d.doc_no, d.doc_name, v.doc_path, v.file_type,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN users u ON o.created_by = u.id
		LEFT JOIN users u2 ON o.updated_by = u2.id
		WHERE o.id = $1
	`

	var doc domain.OutgoingDoc
	var docNo, docName, docPath, fileType, creatorName, updaterName *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.OutgoingNo, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt,
		&docNo, &docName, &docPath, &fileType, &creatorName, &updaterName,
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
	if fileType != nil { doc.FileType = *fileType }
	if creatorName != nil { doc.CreatorName = *creatorName }
	if updaterName != nil { doc.UpdaterName = *updaterName }

	return &doc, nil
}

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int) ([]domain.OutgoingDoc, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT o.id) FROM outgoing_docs o
		JOIN doc_details d ON o.doc_details_id = d.id
		JOIN users u ON o.created_by = u.id
		WHERE u.department_id = $1
	`, deptID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count outgoing documents by department: %w", err)
	}

	query := `
		SELECT
			o.id, o.outgoing_no, o.doc_details_id, o.folder_id, o.created_by, o.updated_by, o.created_at,
			d.doc_no, d.doc_name, v.doc_path, v.file_type,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN users u ON o.created_by = u.id
		LEFT JOIN users u2 ON o.updated_by = u2.id
		WHERE u.department_id = $1
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
		var docNo, docName, docPath, fileType, creatorName, updaterName *string
		if err := rows.Scan(
			&doc.ID, &doc.OutgoingNo, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt,
			&docNo, &docName, &docPath, &fileType, &creatorName, &updaterName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan outgoing document: %w", err)
		}
		if docNo != nil      { doc.DocNo = *docNo }
		if docName != nil    { doc.DocName = *docName }
		if docPath != nil    { doc.DocPath = *docPath }
		if fileType != nil   { doc.FileType = *fileType }
		if creatorName != nil { doc.CreatorName = *creatorName }
		if updaterName != nil { doc.UpdaterName = *updaterName }
		documents = append(documents, doc)
	}
	return documents, total, nil
}

func (r *postgresRepository) FindRecipientsByOutgoingDocID(ctx context.Context, outgoingDocID uuid.UUID) ([]domain.RecipientInfo, error) {
	query := `
		SELECT
			i.dept_id,
			dept.dept_name,
			i.incoming_no,
			i.status,
			i.received_date,
			i.approver_date
		FROM incoming_docs i
		LEFT JOIN departments dept ON i.dept_id = dept.id
		WHERE i.outgoing_doc_id = $1
		ORDER BY dept.dept_name ASC
	`

	rows, err := r.pool.Query(ctx, query, outgoingDocID)
	if err != nil {
		return nil, fmt.Errorf("failed to find recipients: %w", err)
	}
	defer rows.Close()

	var recipients []domain.RecipientInfo
	for rows.Next() {
		var rec domain.RecipientInfo
		var deptID *uuid.UUID
		var deptName, incomingNo *string
		var receivedDate, approverDate *time.Time
		if err := rows.Scan(&deptID, &deptName, &incomingNo, &rec.Status, &receivedDate, &approverDate); err != nil {
			return nil, fmt.Errorf("failed to scan recipient: %w", err)
		}
		if deptID != nil     { rec.DepartmentID = *deptID }
		if deptName != nil   { rec.DepartmentName = *deptName }
		if incomingNo != nil { rec.IncomingNo = *incomingNo }
		rec.ReceivedDate = receivedDate
		rec.ApproverDate = approverDate
		recipients = append(recipients, rec)
	}
	return recipients, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.OutgoingDoc) error {
	query := `
		INSERT INTO outgoing_docs (outgoing_no, doc_details_id, folder_id, created_by, updated_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	doc.CreatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, doc.OutgoingNo, doc.DocDetailsID, doc.FolderID, doc.CreatedBy, doc.UpdatedBy, doc.CreatedAt).
		Scan(&doc.ID)
	if err != nil {
		return fmt.Errorf("failed to create outgoing document: %w", err)
	}

	return nil
}
