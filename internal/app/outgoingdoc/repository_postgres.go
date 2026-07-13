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

// buildDocFilter appends optional filter conditions to a WHERE clause.
// baseCount is the number of positional args already bound (before calling this).
// It returns the WHERE additions string and the new arg values to append.
func buildDocFilter(baseCount int, filter DocFilter) (string, []any) {
	var where string
	var args []any
	n := baseCount
	if filter.DocNo != "" {
		n++
		args = append(args, "%"+filter.DocNo+"%")
		where += fmt.Sprintf(" AND d.doc_no ILIKE $%d", n)
	}
	if filter.Status != "" {
		n++
		args = append(args, filter.Status)
		where += fmt.Sprintf(" AND o.status = $%d", n)
	}
	if filter.StartDate != nil {
		n++
		args = append(args, *filter.StartDate)
		where += fmt.Sprintf(" AND o.created_at >= $%d", n)
	}
	if filter.EndDate != nil {
		n++
		args = append(args, *filter.EndDate)
		where += fmt.Sprintf(" AND o.created_at < $%d", n)
	}
	return where, args
}

func (r *postgresRepository) FindAll(ctx context.Context, limit, offset int, filter DocFilter) ([]domain.OutgoingDoc, int, error) {
	filterWhere, filterArgs := buildDocFilter(0, filter)

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		WHERE o.deleted_at IS NULL%s
	`, filterWhere)
	err := r.pool.QueryRow(ctx, countQuery, filterArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count outgoing documents: %w", err)
	}

	queryArgs := append(filterArgs, limit, offset)
	limitIdx := len(queryArgs) - 1
	offsetIdx := len(queryArgs)
	query := fmt.Sprintf(`
		SELECT
			o.id, o.doc_details_id, o.folder_id, o.owner_dept_id, o.created_by, o.updated_by, o.created_at, o.status,
			d.doc_no, d.doc_name, v.doc_path, v.file_type,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name,
			dept.dept_name as owner_dept_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN users u ON o.created_by = u.id
		LEFT JOIN users u2 ON o.updated_by = u2.id
		LEFT JOIN departments dept ON o.owner_dept_id = dept.id
		WHERE o.deleted_at IS NULL%s
		ORDER BY o.created_at DESC
		LIMIT $%d OFFSET $%d
	`, filterWhere, limitIdx, offsetIdx)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find outgoing documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.OutgoingDoc
	for rows.Next() {
		var doc domain.OutgoingDoc
		var docNo, docName, docPath, fileType, creatorName, updaterName, ownerDeptName *string
		if err := rows.Scan(
			&doc.ID, &doc.DocDetailsID, &doc.FolderID, &doc.OwnerDeptID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt, &doc.Status,
			&docNo, &docName, &docPath, &fileType, &creatorName, &updaterName, &ownerDeptName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan outgoing document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
		if fileType != nil { doc.FileType = *fileType }
		if creatorName != nil { doc.CreatorName = *creatorName }
		if updaterName != nil { doc.UpdaterName = *updaterName }
		if ownerDeptName != nil { doc.OwnerDeptName = *ownerDeptName }

		documents = append(documents, doc)
	}

	return documents, total, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDoc, error) {
	query := `
		SELECT
			o.id, o.doc_details_id, o.folder_id, o.owner_dept_id, o.created_by, o.updated_by, o.created_at, o.status,
			d.doc_no, d.doc_name, v.doc_path, v.file_type,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name,
			dept.dept_name as owner_dept_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN users u ON o.created_by = u.id
		LEFT JOIN users u2 ON o.updated_by = u2.id
		LEFT JOIN departments dept ON o.owner_dept_id = dept.id
		WHERE o.id = $1 AND o.deleted_at IS NULL
	`

	var doc domain.OutgoingDoc
	var docNo, docName, docPath, fileType, creatorName, updaterName, ownerDeptName *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.DocDetailsID, &doc.FolderID, &doc.OwnerDeptID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt, &doc.Status,
		&docNo, &docName, &docPath, &fileType, &creatorName, &updaterName, &ownerDeptName,
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
	if ownerDeptName != nil { doc.OwnerDeptName = *ownerDeptName }

	return &doc, nil
}

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int, filter DocFilter) ([]domain.OutgoingDoc, int, error) {
	// $1 is always deptID; filter args start at $2
	filterWhere, extraArgs := buildDocFilter(1, filter)
	baseArgs := append([]any{deptID}, extraArgs...)

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT o.id) FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		JOIN users u ON o.created_by = u.id
		WHERE u.department_id = $1 AND o.deleted_at IS NULL%s
	`, filterWhere)
	err := r.pool.QueryRow(ctx, countQuery, baseArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count outgoing documents by department: %w", err)
	}

	queryArgs := append(baseArgs, limit, offset)
	limitIdx := len(queryArgs) - 1
	offsetIdx := len(queryArgs)
	query := fmt.Sprintf(`
		SELECT
			o.id, o.doc_details_id, o.folder_id, o.owner_dept_id, o.created_by, o.updated_by, o.created_at, o.status,
			d.doc_no, d.doc_name, v.doc_path, v.file_type,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name,
			dept.dept_name as owner_dept_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN users u ON o.created_by = u.id
		LEFT JOIN users u2 ON o.updated_by = u2.id
		LEFT JOIN departments dept ON o.owner_dept_id = dept.id
		WHERE u.department_id = $1 AND o.deleted_at IS NULL%s
		ORDER BY o.created_at DESC
		LIMIT $%d OFFSET $%d
	`, filterWhere, limitIdx, offsetIdx)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find outgoing documents by department: %w", err)
	}
	defer rows.Close()

	var documents []domain.OutgoingDoc
	for rows.Next() {
		var doc domain.OutgoingDoc
		var docNo, docName, docPath, fileType, creatorName, updaterName, ownerDeptName *string
		if err := rows.Scan(
			&doc.ID, &doc.DocDetailsID, &doc.FolderID, &doc.OwnerDeptID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt, &doc.Status,
			&docNo, &docName, &docPath, &fileType, &creatorName, &updaterName, &ownerDeptName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan outgoing document: %w", err)
		}
		if docNo != nil       { doc.DocNo = *docNo }
		if docName != nil     { doc.DocName = *docName }
		if docPath != nil     { doc.DocPath = *docPath }
		if fileType != nil    { doc.FileType = *fileType }
		if creatorName != nil { doc.CreatorName = *creatorName }
		if updaterName != nil { doc.UpdaterName = *updaterName }
		if ownerDeptName != nil { doc.OwnerDeptName = *ownerDeptName }
		documents = append(documents, doc)
	}
	return documents, total, nil
}

// FindRecipientsByOutgoingDocID builds the ordered recipient list from the
// route table LEFT JOINed with incoming_docs. Steps whose incoming doc has not
// been created yet appear with status "waiting".
func (r *postgresRepository) FindRecipientsByOutgoingDocID(ctx context.Context, outgoingDocID uuid.UUID) ([]domain.RecipientInfo, error) {
	query := `
		SELECT
			rt.dept_id,
			dept.dept_name,
			rt.sequence_order,
			i.status,
			i.received_date,
			i.approver_date
		FROM outgoing_doc_routes rt
		LEFT JOIN departments dept ON rt.dept_id = dept.id
		LEFT JOIN incoming_docs i ON rt.incoming_doc_id = i.id
		WHERE rt.outgoing_doc_id = $1
		ORDER BY rt.sequence_order ASC
	`

	rows, err := r.pool.Query(ctx, query, outgoingDocID)
	if err != nil {
		return nil, fmt.Errorf("failed to find recipients: %w", err)
	}
	defer rows.Close()

	var recipients []domain.RecipientInfo
	currentMarked := false
	for rows.Next() {
		var rec domain.RecipientInfo
		var deptID *uuid.UUID
		var deptName, status *string
		var receivedDate, approverDate *time.Time
		if err := rows.Scan(&deptID, &deptName, &rec.SequenceOrder, &status, &receivedDate, &approverDate); err != nil {
			return nil, fmt.Errorf("failed to scan recipient: %w", err)
		}
		if deptID != nil {
			rec.DepartmentID = *deptID
		}
		if deptName != nil {
			rec.DepartmentName = *deptName
		}
		if status != nil {
			rec.Status = *status
		} else {
			rec.Status = domain.StatusWaiting
		}
		rec.ReceivedDate = receivedDate
		rec.ApproverDate = approverDate
		// The current step is the first step that is not yet approved/rejected.
		if !currentMarked && rec.Status != string(domain.IncomingStatusApproved) && rec.Status != string(domain.IncomingStatusRejected) {
			rec.IsCurrent = true
			currentMarked = true
		}
		recipients = append(recipients, rec)
	}
	return recipients, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.OutgoingDoc) error {
	query := `
		INSERT INTO outgoing_docs (doc_details_id, folder_id, owner_dept_id, created_by, updated_by, created_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	doc.CreatedAt = time.Now()
	if doc.Status == "" {
		doc.Status = domain.OutgoingStatusPending
	}

	err := r.pool.QueryRow(ctx, query, doc.DocDetailsID, doc.FolderID, doc.OwnerDeptID, doc.CreatedBy, doc.UpdatedBy, doc.CreatedAt, doc.Status).
		Scan(&doc.ID)
	if err != nil {
		return fmt.Errorf("failed to create outgoing document: %w", err)
	}

	return nil
}

// UpdateDocMeta patches doc_no / doc_name on the linked doc_details, and bumps
// outgoing_docs.updated_by. Skips fields whose value is an empty string.
// Runs inside a transaction so a partial write does not leave the two tables out of sync.
func (r *postgresRepository) UpdateDocMeta(ctx context.Context, id uuid.UUID, docNo, docName string, updaterID *uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if docNo != "" || docName != "" {
		// Update only the columns that are non-empty. COALESCE(NULLIF(...))
		// keeps the original when the argument is empty.
		_, err = tx.Exec(ctx, `
			UPDATE doc_details
			SET
				doc_no   = COALESCE(NULLIF($1, ''), doc_no),
				doc_name = COALESCE(NULLIF($2, ''), doc_name),
				updated_at = NOW()
			WHERE id = (SELECT doc_details_id FROM outgoing_docs WHERE id = $3)
		`, docNo, docName, id)
		if err != nil {
			return fmt.Errorf("update doc_details: %w", err)
		}
	}

	_, err = tx.Exec(ctx,
		`UPDATE outgoing_docs SET updated_by = COALESCE($1, updated_by) WHERE id = $2`,
		updaterID, id,
	)
	if err != nil {
		return fmt.Errorf("update outgoing_docs.updated_by: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// SoftDelete marks outgoing_docs and its linked doc_details as deleted.
// The two writes run in a transaction so the pair stays consistent.
func (r *postgresRepository) SoftDelete(ctx context.Context, id uuid.UUID, updaterID *uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx,
		`UPDATE outgoing_docs SET deleted_at = NOW(), updated_by = COALESCE($1, updated_by) WHERE id = $2`,
		updaterID, id,
	); err != nil {
		return fmt.Errorf("soft-delete outgoing_docs: %w", err)
	}

	if _, err = tx.Exec(ctx, `
		UPDATE doc_details SET deleted_at = NOW()
		WHERE id = (SELECT doc_details_id FROM outgoing_docs WHERE id = $1)
	`, id); err != nil {
		return fmt.Errorf("soft-delete doc_details: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// ReplaceRoutes clears the existing route steps and rewrites them from deptIDs
// in order. Meant to be called only while the doc is still pending — otherwise
// existing incoming_docs would be silently orphaned.
func (r *postgresRepository) ReplaceRoutes(ctx context.Context, outgoingDocID uuid.UUID, deptIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx,
		`DELETE FROM outgoing_doc_routes WHERE outgoing_doc_id = $1`, outgoingDocID,
	); err != nil {
		return fmt.Errorf("delete existing routes: %w", err)
	}

	for i, deptID := range deptIDs {
		if _, err = tx.Exec(ctx, `
			INSERT INTO outgoing_doc_routes (outgoing_doc_id, dept_id, sequence_order)
			VALUES ($1, $2, $3)
		`, outgoingDocID, deptID, i+1); err != nil {
			return fmt.Errorf("insert route %d: %w", i+1, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// FindByDocDetailsID returns the outgoing_doc row backed by a given doc_details.
// Only considers non-deleted rows.
func (r *postgresRepository) FindByDocDetailsID(ctx context.Context, docDetailsID uuid.UUID) (*domain.OutgoingDoc, error) {
	var doc domain.OutgoingDoc
	err := r.pool.QueryRow(ctx, `
		SELECT id, doc_details_id, folder_id, owner_dept_id, created_by, updated_by, created_at, status
		FROM outgoing_docs
		WHERE doc_details_id = $1 AND deleted_at IS NULL
	`, docDetailsID).Scan(
		&doc.ID, &doc.DocDetailsID, &doc.FolderID, &doc.OwnerDeptID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt, &doc.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find by doc_details_id: %w", err)
	}
	return &doc, nil
}

// ReplaceFile swaps the version row's stored file (doc_path + file_type) for the
// outgoing doc's doc_details. Runs in a tx so a partial write can't leave the
// version pointing at a nonexistent file with an updated updated_by.
func (r *postgresRepository) ReplaceFile(ctx context.Context, outgoingDocID uuid.UUID, docPath, fileType string, updaterID *uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE versions
		SET doc_path = $1, file_type = $2
		WHERE doc_details_id = (SELECT doc_details_id FROM outgoing_docs WHERE id = $3)
	`, docPath, fileType, outgoingDocID)
	if err != nil {
		return fmt.Errorf("update versions: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no version row found for outgoing doc %s", outgoingDocID)
	}

	if _, err = tx.Exec(ctx,
		`UPDATE outgoing_docs SET updated_by = COALESCE($1, updated_by) WHERE id = $2`,
		updaterID, outgoingDocID,
	); err != nil {
		return fmt.Errorf("bump updated_by: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// UpdateStatus sets the owner-approval gate status and records the approver.
func (r *postgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, approverID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE outgoing_docs SET status = $1, updated_by = COALESCE($2, updated_by) WHERE id = $3`,
		status, approverID, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update outgoing document status: %w", err)
	}
	return nil
}

// CreateRoutes inserts the ordered route steps for an outgoing document and
// writes the generated ids back into the passed slice.
func (r *postgresRepository) CreateRoutes(ctx context.Context, routes []domain.OutgoingDocRoute) error {
	if len(routes) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, rt := range routes {
		batch.Queue(
			`INSERT INTO outgoing_doc_routes (outgoing_doc_id, dept_id, sequence_order, incoming_doc_id)
			 VALUES ($1, $2, $3, $4) RETURNING id`,
			rt.OutgoingDocID, rt.DeptID, rt.SequenceOrder, rt.IncomingDocID,
		)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := range routes {
		if err := br.QueryRow().Scan(&routes[i].ID); err != nil {
			return fmt.Errorf("failed to create route step: %w", err)
		}
	}
	return nil
}

// FindRouteByIncomingDocID returns the route step linked to a given incoming doc.
func (r *postgresRepository) FindRouteByIncomingDocID(ctx context.Context, incomingDocID uuid.UUID) (*domain.RouteStep, error) {
	query := `
		SELECT id, outgoing_doc_id, dept_id, sequence_order, incoming_doc_id
		FROM outgoing_doc_routes
		WHERE incoming_doc_id = $1
	`
	var step domain.RouteStep
	err := r.pool.QueryRow(ctx, query, incomingDocID).
		Scan(&step.ID, &step.OutgoingDocID, &step.DeptID, &step.SequenceOrder, &step.IncomingDocID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find route by incoming doc: %w", err)
	}
	return &step, nil
}

// FindNextStep returns the route step immediately after currentOrder, or nil if none.
func (r *postgresRepository) FindNextStep(ctx context.Context, outgoingDocID uuid.UUID, currentOrder int) (*domain.RouteStep, error) {
	query := `
		SELECT id, outgoing_doc_id, dept_id, sequence_order, incoming_doc_id
		FROM outgoing_doc_routes
		WHERE outgoing_doc_id = $1 AND sequence_order > $2
		ORDER BY sequence_order ASC
		LIMIT 1
	`
	var step domain.RouteStep
	err := r.pool.QueryRow(ctx, query, outgoingDocID, currentOrder).
		Scan(&step.ID, &step.OutgoingDocID, &step.DeptID, &step.SequenceOrder, &step.IncomingDocID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find next route step: %w", err)
	}
	return &step, nil
}

// AttachIncomingDoc links a created incoming doc to its route step.
func (r *postgresRepository) AttachIncomingDoc(ctx context.Context, routeID, incomingDocID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE outgoing_doc_routes SET incoming_doc_id = $1 WHERE id = $2`,
		incomingDocID, routeID,
	)
	if err != nil {
		return fmt.Errorf("failed to attach incoming doc to route: %w", err)
	}
	return nil
}
