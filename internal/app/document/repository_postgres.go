package document

import (
	"context"
	"database/sql"
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

type joinedDocScan struct {
	domain.DocDetails
	DocTypeName *string
	UserName    *string
	FileType    *string
}

func buildDocResponse(row *joinedDocScan) domain.DocumentResponse {
	resp := row.DocDetails.ToResponse()
	if row.DocTypeName != nil {
		resp.DocTypeName = *row.DocTypeName
	}
	if row.UserName != nil {
		resp.UserName = *row.UserName
	}
	if row.FileType != nil {
		resp.FileType = *row.FileType
	}
	return resp
}

const joinedDocSelect = `
	SELECT
		d.id, d.doc_no, d.doc_name, d.description, d.version_number, d.status,
		d.doc_type_id, d.user_id, d.created_at, d.updated_at, d.deleted_at,
		dt.type_name AS doc_type_name,
		u.firstname || ' ' || u.lastname AS user_name,
		v.file_type
	FROM doc_details d
	LEFT JOIN doc_types dt ON dt.id = d.doc_type_id
	LEFT JOIN users u ON u.id = d.user_id
	LEFT JOIN versions v ON v.doc_details_id = d.id AND v.version_number = d.version_number
`

func scanJoinedDoc(row pgx.Row) (*joinedDocScan, error) {
	var s joinedDocScan
	var desc sql.NullString
	var deletedAt sql.NullTime

	err := row.Scan(
		&s.ID, &s.DocNo, &s.DocName, &desc, &s.VersionNumber, &s.Status,
		&s.DocTypeID, &s.UserID, &s.CreatedAt, &s.UpdatedAt, &deletedAt,
		&s.DocTypeName, &s.UserName, &s.FileType,
	)
	if desc.Valid {
		s.Description = &desc.String
	}
	if deletedAt.Valid {
		s.DeletedAt = &deletedAt.Time
	}
	return &s, err
}

func (r *postgresRepository) FindAllJoined(ctx context.Context) ([]domain.DocumentResponse, error) {
	query := joinedDocSelect + `
		WHERE d.deleted_at IS NULL
		  AND v.folder_id IS NULL
		  AND NOT EXISTS (SELECT 1 FROM outgoing_docs o WHERE o.doc_details_id = d.id)
		  AND NOT EXISTS (SELECT 1 FROM incoming_docs i WHERE i.doc_details_id = d.id)
		ORDER BY d.created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer rows.Close()

	var responses []domain.DocumentResponse
	for rows.Next() {
		var s joinedDocScan
		var desc sql.NullString
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&s.ID, &s.DocNo, &s.DocName, &desc, &s.VersionNumber, &s.Status,
			&s.DocTypeID, &s.UserID, &s.CreatedAt, &s.UpdatedAt, &deletedAt,
			&s.DocTypeName, &s.UserName, &s.FileType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		if desc.Valid {
			s.Description = &desc.String
		}
		if deletedAt.Valid {
			s.DeletedAt = &deletedAt.Time
		}
		responses = append(responses, buildDocResponse(&s))
	}

	return responses, nil
}

func (r *postgresRepository) FindByIDJoined(ctx context.Context, id uuid.UUID) (*domain.DocumentResponse, error) {
	query := joinedDocSelect + ` WHERE d.id = $1 AND d.deleted_at IS NULL`

	s, err := scanJoinedDoc(r.pool.QueryRow(ctx, query, id))
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}
	resp := buildDocResponse(s)
	return &resp, nil
}

func (r *postgresRepository) FindByFolderIDJoined(ctx context.Context, folderID uuid.UUID) ([]domain.DocumentResponse, error) {
	query := joinedDocSelect + `
		WHERE v.folder_id = $1
		  AND d.deleted_at IS NULL
		  AND NOT EXISTS (SELECT 1 FROM outgoing_docs o WHERE o.doc_details_id = d.id)
		  AND NOT EXISTS (SELECT 1 FROM incoming_docs i WHERE i.doc_details_id = d.id)
		ORDER BY d.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents by folder: %w", err)
	}
	defer rows.Close()

	var responses []domain.DocumentResponse
	for rows.Next() {
		var s joinedDocScan
		var desc sql.NullString
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&s.ID, &s.DocNo, &s.DocName, &desc, &s.VersionNumber, &s.Status,
			&s.DocTypeID, &s.UserID, &s.CreatedAt, &s.UpdatedAt, &deletedAt,
			&s.DocTypeName, &s.UserName, &s.FileType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		if desc.Valid {
			s.Description = &desc.String
		}
		if deletedAt.Valid {
			s.DeletedAt = &deletedAt.Time
		}
		responses = append(responses, buildDocResponse(&s))
	}

	return responses, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.DocDetails, error) {
	query := `
		SELECT id, doc_no, doc_name, description, version_number, status, doc_type_id, user_id, created_at, updated_at, deleted_at
		FROM doc_details
		WHERE id = $1 AND deleted_at IS NULL
	`

	var doc domain.DocDetails
	var desc sql.NullString
	var deletedAt sql.NullTime
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.DocNo, &doc.DocName, &desc, &doc.VersionNumber, &doc.Status,
		&doc.DocTypeID, &doc.UserID, &doc.CreatedAt, &doc.UpdatedAt, &deletedAt,
	)
	if desc.Valid {
		doc.Description = &desc.String
	}
	if deletedAt.Valid {
		doc.DeletedAt = &deletedAt.Time
	}
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	return &doc, nil
}

func (r *postgresRepository) FindByDocNo(ctx context.Context, docNo string) (*domain.DocDetails, error) {
	query := `
		SELECT id, doc_no, doc_name, description, version_number, status, doc_type_id, user_id, created_at, updated_at, deleted_at
		FROM doc_details
		WHERE doc_no = $1 AND deleted_at IS NULL
	`

	var doc domain.DocDetails
	var desc sql.NullString
	var deletedAt sql.NullTime
	err := r.pool.QueryRow(ctx, query, docNo).Scan(
		&doc.ID, &doc.DocNo, &doc.DocName, &desc, &doc.VersionNumber, &doc.Status,
		&doc.DocTypeID, &doc.UserID, &doc.CreatedAt, &doc.UpdatedAt, &deletedAt,
	)
	if desc.Valid {
		doc.Description = &desc.String
	}
	if deletedAt.Valid {
		doc.DeletedAt = &deletedAt.Time
	}
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find document by doc_no: %w", err)
	}

	return &doc, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.DocDetails) error {
	query := `
		INSERT INTO doc_details (doc_no, doc_name, description, version_number, status, doc_type_id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query,
		doc.DocNo, doc.DocName, doc.Description, doc.VersionNumber, doc.Status, doc.DocTypeID, doc.UserID, doc.CreatedAt, doc.UpdatedAt,
	).Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id uuid.UUID, doc *domain.DocDetails) error {
	query := `
		UPDATE doc_details 
		SET doc_name = $1, description = $2, version_number = $3, status = $4, doc_type_id = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	doc.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query,
		doc.DocName, doc.Description, doc.VersionNumber, doc.Status, doc.DocTypeID, doc.UpdatedAt, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE doc_details SET deleted_at = NOW() WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

func (r *postgresRepository) CreateVersion(ctx context.Context, version *domain.Version) error {
	query := `
		INSERT INTO versions (doc_details_id, folder_id, version_number, doc_path, file_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	version.CreatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, version.DocDetailsID, version.FolderID, version.VersionNumber, version.DocPath, version.FileType, version.CreatedAt).
		Scan(&version.ID, &version.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetVersionsByDocID(ctx context.Context, docDetailsID uuid.UUID) ([]domain.Version, error) {
	query := `
		SELECT id, doc_details_id, folder_id, version_number, doc_path, file_type, created_at
		FROM versions
		WHERE doc_details_id = $1 AND deleted_at IS NULL
		ORDER BY version_number DESC
	`

	rows, err := r.pool.Query(ctx, query, docDetailsID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	defer rows.Close()

	var versions []domain.Version
	for rows.Next() {
		var v domain.Version
		if err := rows.Scan(&v.ID, &v.DocDetailsID, &v.FolderID, &v.VersionNumber, &v.DocPath, &v.FileType, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}
