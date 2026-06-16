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
		args = append(args, filter.DocNo)
		where += fmt.Sprintf(" AND d.doc_no = $%d", n)
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
		WHERE 1=1%s
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
			o.id, o.doc_details_id, o.folder_id, o.created_by, o.updated_by, o.created_at, o.status,
			d.doc_no, d.doc_name, v.doc_path, v.file_type,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN users u ON o.created_by = u.id
		LEFT JOIN users u2 ON o.updated_by = u2.id
		WHERE 1=1%s
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
		var docNo, docName, docPath, fileType, creatorName, updaterName *string
		if err := rows.Scan(
			&doc.ID, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt, &doc.Status,
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
			o.id, o.doc_details_id, o.folder_id, o.created_by, o.updated_by, o.created_at, o.status,
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
		&doc.ID, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt, &doc.Status,
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

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int, filter DocFilter) ([]domain.OutgoingDoc, int, error) {
	// $1 is always deptID; filter args start at $2
	filterWhere, extraArgs := buildDocFilter(1, filter)
	baseArgs := append([]any{deptID}, extraArgs...)

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT o.id) FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		JOIN users u ON o.created_by = u.id
		WHERE u.department_id = $1%s
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
			o.id, o.doc_details_id, o.folder_id, o.created_by, o.updated_by, o.created_at, o.status,
			d.doc_no, d.doc_name, v.doc_path, v.file_type,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN users u ON o.created_by = u.id
		LEFT JOIN users u2 ON o.updated_by = u2.id
		WHERE u.department_id = $1%s
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
		var docNo, docName, docPath, fileType, creatorName, updaterName *string
		if err := rows.Scan(
			&doc.ID, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt, &doc.Status,
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
		INSERT INTO outgoing_docs (doc_details_id, folder_id, created_by, updated_by, created_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	doc.CreatedAt = time.Now()
	if doc.Status == "" {
		doc.Status = domain.OutgoingStatusPending
	}

	err := r.pool.QueryRow(ctx, query, doc.DocDetailsID, doc.FolderID, doc.CreatedBy, doc.UpdatedBy, doc.CreatedAt, doc.Status).
		Scan(&doc.ID)
	if err != nil {
		return fmt.Errorf("failed to create outgoing document: %w", err)
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
