package rejecteddoc

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new Postgres repository for rejected documents.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) FindRejected(ctx context.Context, deptID *uuid.UUID, start, end *time.Time, limit, offset int) ([]domain.RejectedDocResponse, int, error) {
	var args []any
	argn := 0
	next := func(v any) string {
		argn++
		args = append(args, v)
		return fmt.Sprintf("$%d", argn)
	}

	// Inbound leg: incoming docs rejected by a receiving department head.
	inboundWhere := "WHERE i.status = 'rejected'"
	if deptID != nil {
		inboundWhere += " AND i.dept_id = " + next(*deptID)
	}
	if start != nil {
		inboundWhere += " AND i.approver_date >= " + next(*start)
	}
	if end != nil {
		inboundWhere += " AND i.approver_date < " + next(*end)
	}
	inbound := fmt.Sprintf(`
		SELECT 'inbound' AS source, i.id, d.doc_no, d.doc_name, v.doc_path, v.file_type,
			i.dept_id, dept.dept_name,
			i.approver_id AS rejected_by_id,
			u.firstname || ' ' || u.lastname AS rejected_by,
			i.approver_date AS rejected_at, i.remark
		FROM incoming_docs i
		LEFT JOIN doc_details d ON i.doc_details_id = d.id
		LEFT JOIN versions v ON i.doc_details_id = v.doc_details_id AND i.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN departments dept ON i.dept_id = dept.id
		LEFT JOIN users u ON i.approver_id = u.id
		%s`, inboundWhere)

	// Outbound leg: outgoing docs our department created that were rejected —
	// either by the owner department head (o.status set directly) or by a
	// recipient department (propagated via ApproveDocument → UpdateStatus).
	// We LEFT JOIN incoming_docs to surface the recipient's rejection details
	// (dept, approver, date, remark) when available; otherwise fall back to
	// the owner-head rejection data stored on outgoing_docs itself.
	outboundWhere := "WHERE o.status = 'rejected'"
	if deptID != nil {
		outboundWhere += " AND o.owner_dept_id = " + next(*deptID)
	}
	if start != nil {
		outboundWhere += " AND COALESCE(ri.approver_date, o.created_at) >= " + next(*start)
	}
	if end != nil {
		outboundWhere += " AND COALESCE(ri.approver_date, o.created_at) < " + next(*end)
	}
	outbound := fmt.Sprintf(`
		SELECT 'outbound' AS source, o.id, d.doc_no, d.doc_name, v.doc_path, v.file_type,
			COALESCE(ri.dept_id, o.owner_dept_id) AS dept_id,
			COALESCE(rd.dept_name, od.dept_name)  AS dept_name,
			COALESCE(ri.approver_id, o.updated_by) AS rejected_by_id,
			COALESCE(rbu.firstname || ' ' || rbu.lastname, obu.firstname || ' ' || obu.lastname) AS rejected_by,
			COALESCE(ri.approver_date, o.created_at) AS rejected_at,
			COALESCE(ri.remark, '') AS remark
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id IS NOT DISTINCT FROM v.folder_id
		LEFT JOIN departments od ON o.owner_dept_id = od.id
		LEFT JOIN incoming_docs ri ON ri.outgoing_doc_id = o.id AND ri.status = 'rejected'
		LEFT JOIN departments rd ON ri.dept_id = rd.id
		LEFT JOIN users rbu ON ri.approver_id = rbu.id
		LEFT JOIN users obu ON o.updated_by = obu.id
		%s`, outboundWhere)

	union := "(" + inbound + ") UNION ALL (" + outbound + ")"

	// Count first — uses only the filter args bound so far (before limit/offset).
	var total int
	countQuery := "SELECT COUNT(*) FROM (" + union + ") AS rejected"
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count rejected documents: %w", err)
	}

	limitPlaceholder := next(limit)
	offsetPlaceholder := next(offset)
	query := union + " ORDER BY rejected_at DESC NULLS LAST LIMIT " + limitPlaceholder + " OFFSET " + offsetPlaceholder

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find rejected documents: %w", err)
	}
	defer rows.Close()

	var docs []domain.RejectedDocResponse
	for rows.Next() {
		var rec domain.RejectedDocResponse
		var docNo, docName, docPath, fileType, deptName, rejectedBy, remark *string
		var deptID, rejectedByID *uuid.UUID
		var rejectedAt *time.Time
		if err := rows.Scan(
			&rec.Source, &rec.ID, &docNo, &docName, &docPath, &fileType,
			&deptID, &deptName, &rejectedByID, &rejectedBy, &rejectedAt, &remark,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan rejected document: %w", err)
		}
		if docNo != nil {
			rec.DocNo = *docNo
		}
		if docName != nil {
			rec.DocName = *docName
		}
		if docPath != nil {
			rec.DocPath = *docPath
		}
		if fileType != nil {
			rec.FileType = *fileType
		}
		rec.DeptID = deptID
		if deptName != nil {
			rec.DeptName = *deptName
		}
		rec.RejectedByID = rejectedByID
		if rejectedBy != nil {
			rec.RejectedBy = *rejectedBy
		}
		rec.RejectedAt = rejectedAt
		if remark != nil {
			rec.Remark = *remark
		}
		docs = append(docs, rec)
	}
	return docs, total, nil
}
