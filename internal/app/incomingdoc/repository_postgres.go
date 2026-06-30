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

// scanIncomingDoc scans a row into IncomingDoc (includes dept_id, outgoing_doc_id + dept_name join)
func scanIncomingDoc(row interface {
	Scan(dest ...any) error
}, doc *domain.IncomingDoc) error {
	var docNo, docName, docPath, fileType, creatorName, updaterName, approverName, deptName *string
	err := row.Scan(
		&doc.ID, &doc.IncomingDate, &doc.ReceivedDate, &doc.Status,
		&doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.ApproverID,
		&doc.ApproverDate, &doc.Remark, &doc.UpdatedAt, &doc.DeptID, &doc.OutgoingDocID,
		&docNo, &docName, &docPath, &fileType, &creatorName, &updaterName, &approverName, &deptName,
	)
	if err != nil {
		return err
	}
	if docNo != nil        { doc.DocNo = *docNo }
	if docName != nil      { doc.DocName = *docName }
	if docPath != nil      { doc.DocPath = *docPath }
	if fileType != nil     { doc.FileType = *fileType }
	if creatorName != nil  { doc.CreatorName = *creatorName }
	if updaterName != nil  { doc.UpdaterName = *updaterName }
	if approverName != nil { doc.ApproverName = *approverName }
	if deptName != nil     { doc.DeptName = *deptName }
	return nil
}

const baseSelectIncoming = `
	SELECT
		i.id, i.incoming_date, i.received_date, i.status,
		i.doc_details_id, i.folder_id, i.created_by, i.updated_by, i.approver_id,
		i.approver_date, i.remark, i.updated_at, i.dept_id, i.outgoing_doc_id,
		d.doc_no, d.doc_name,
		v.doc_path, v.file_type,
		u1.firstname || ' ' || u1.lastname as creator_name,
		u2.firstname || ' ' || u2.lastname as updater_name,
		u3.firstname || ' ' || u3.lastname as approver_name,
		dept.dept_name
	FROM incoming_docs i
	LEFT JOIN doc_details d ON i.doc_details_id = d.id
	LEFT JOIN versions v ON i.doc_details_id = v.doc_details_id AND i.folder_id IS NOT DISTINCT FROM v.folder_id
	LEFT JOIN users u1 ON i.created_by = u1.id
	LEFT JOIN users u2 ON i.updated_by = u2.id
	LEFT JOIN users u3 ON i.approver_id = u3.id
	LEFT JOIN departments dept ON i.dept_id = dept.id`

func (r *postgresRepository) FindAll(ctx context.Context, limit, offset int) ([]domain.IncomingDoc, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM incoming_docs").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count incoming documents: %w", err)
	}

	query := baseSelectIncoming + `
		ORDER BY i.updated_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find incoming documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := scanIncomingDoc(rows, &doc); err != nil {
			return nil, 0, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}
	return documents, total, nil
}

func (r *postgresRepository) FindAllExcludingSender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.IncomingDoc, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM incoming_docs WHERE created_by != $1", senderID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count incoming documents: %w", err)
	}

	query := baseSelectIncoming + `
		WHERE i.created_by != $1
		ORDER BY i.updated_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, senderID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find incoming documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := scanIncomingDoc(rows, &doc); err != nil {
			return nil, 0, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}
	return documents, total, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDoc, error) {
	query := baseSelectIncoming + ` WHERE i.id = $1`

	var doc domain.IncomingDoc
	err := scanIncomingDoc(r.pool.QueryRow(ctx, query, id), &doc)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("incoming document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming document: %w", err)
	}
	return &doc, nil
}

func (r *postgresRepository) FindByReceiverID(ctx context.Context, receiverID uuid.UUID) ([]domain.IncomingDoc, error) {
	query := baseSelectIncoming + `
		WHERE i.updated_by = $1
		ORDER BY i.updated_at DESC`

	rows, err := r.pool.Query(ctx, query, receiverID)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by receiver: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := scanIncomingDoc(rows, &doc); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int) ([]domain.IncomingDoc, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM incoming_docs WHERE dept_id = $1", deptID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count incoming documents by department: %w", err)
	}

	query := baseSelectIncoming + `
		WHERE i.dept_id = $1
		ORDER BY i.updated_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, deptID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find incoming documents by department: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := scanIncomingDoc(rows, &doc); err != nil {
			return nil, 0, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}
	return documents, total, nil
}

func (r *postgresRepository) FindByStatus(ctx context.Context, status string) ([]domain.IncomingDoc, error) {
	query := baseSelectIncoming + `
		WHERE i.status = $1
		ORDER BY i.updated_at DESC`

	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by status: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := scanIncomingDoc(rows, &doc); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func (r *postgresRepository) FindByDocID(ctx context.Context, docDetailsID uuid.UUID) ([]domain.IncomingDoc, error) {
	query := baseSelectIncoming + `
		WHERE i.doc_details_id = $1
		ORDER BY i.updated_at DESC`

	rows, err := r.pool.Query(ctx, query, docDetailsID)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming documents by doc details id: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := scanIncomingDoc(rows, &doc); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func (r *postgresRepository) FindByDocNo(ctx context.Context, docNo string) (*domain.IncomingDoc, error) {
	query := baseSelectIncoming + ` WHERE d.doc_no = $1 ORDER BY i.updated_at DESC LIMIT 1`

	var doc domain.IncomingDoc
	err := scanIncomingDoc(r.pool.QueryRow(ctx, query, docNo), &doc)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("incoming document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming document by doc_no: %w", err)
	}
	return &doc, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.IncomingDoc) error {
	query := `
		INSERT INTO incoming_docs (
			incoming_date, received_date, status,
			doc_details_id, folder_id, created_by, updated_by,
			approver_id, approver_date, remark, updated_at, dept_id, outgoing_doc_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	doc.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query,
		doc.IncomingDate, doc.ReceivedDate, doc.Status,
		doc.DocDetailsID, doc.FolderID, doc.CreatedBy, doc.UpdatedBy,
		doc.ApproverID, doc.ApproverDate, doc.Remark, doc.UpdatedAt,
		doc.DeptID, doc.OutgoingDocID,
	).Scan(&doc.ID)
	if err != nil {
		return fmt.Errorf("failed to create incoming document: %w", err)
	}
	return nil
}

func (r *postgresRepository) FindByOutgoingDocID(ctx context.Context, outgoingDocID uuid.UUID) ([]domain.IncomingDoc, error) {
	query := baseSelectIncoming + `
		WHERE i.outgoing_doc_id = $1
		ORDER BY dept.dept_name ASC`

	rows, err := r.pool.Query(ctx, query, outgoingDocID)
	if err != nil {
		return nil, fmt.Errorf("failed to find incoming docs by outgoing_doc_id: %w", err)
	}
	defer rows.Close()

	var documents []domain.IncomingDoc
	for rows.Next() {
		var doc domain.IncomingDoc
		if err := scanIncomingDoc(rows, &doc); err != nil {
			return nil, fmt.Errorf("failed to scan incoming document: %w", err)
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func (r *postgresRepository) Update(ctx context.Context, id uuid.UUID, doc *domain.IncomingDoc) error {
	query := `
		UPDATE incoming_docs
		SET updated_by = $1, approver_id = $2,
		    received_date = $3, approver_date = $4, remark = $5,
		    status = $6, updated_at = $7
		WHERE id = $8`
	doc.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query,
		doc.UpdatedBy, doc.ApproverID,
		doc.ReceivedDate, doc.ApproverDate, doc.Remark,
		doc.Status, doc.UpdatedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update incoming document: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("incoming document not found")
	}
	return nil
}

func (r *postgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE incoming_docs SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update incoming document status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("incoming document not found")
	}
	return nil
}
