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
			d.doc_no, d.doc_name, v.doc_path,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id = v.folder_id
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
		var docNo, docName, docPath, creatorName, updaterName *string
		if err := rows.Scan(
			&doc.ID, &doc.OutgoingNo, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt,
			&docNo, &docName, &docPath, &creatorName, &updaterName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan outgoing document: %w", err)
		}

		if docNo != nil { doc.DocNo = *docNo }
		if docName != nil { doc.DocName = *docName }
		if docPath != nil { doc.DocPath = *docPath }
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
			d.doc_no, d.doc_name, v.doc_path,
			u.firstname || ' ' || u.lastname as creator_name,
			u2.firstname || ' ' || u2.lastname as updater_name
		FROM outgoing_docs o
		LEFT JOIN doc_details d ON o.doc_details_id = d.id
		LEFT JOIN versions v ON o.doc_details_id = v.doc_details_id AND o.folder_id = v.folder_id
		LEFT JOIN users u ON o.created_by = u.id
		LEFT JOIN users u2 ON o.updated_by = u2.id
		WHERE o.id = $1
	`

	var doc domain.OutgoingDoc
	var docNo, docName, docPath, creatorName, updaterName *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.OutgoingNo, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy, &doc.UpdatedBy, &doc.CreatedAt,
		&docNo, &docName, &docPath, &creatorName, &updaterName,
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
	if creatorName != nil { doc.CreatorName = *creatorName }
	if updaterName != nil { doc.UpdaterName = *updaterName }

	return &doc, nil
}

func (r *postgresRepository) FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int) ([]domain.OutgoingDoc, int, error) {
	return []domain.OutgoingDoc{}, 0, nil
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
