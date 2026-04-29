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

// NewPostgresRepository creates a new Postgres repository
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

// joinedDocScan is a helper struct for scanning JOIN results
type joinedDocScan struct {
	domain.Document
	DocTypeName     *string
	FolderName      *string
	RegistrantFirst *string
	RegistrantLast  *string
	RegistrantEmail *string
	DepartmentName  *string
	SectorName      *string
}

func buildDocResponse(row *joinedDocScan) domain.DocumentResponse {
	resp := row.Document.ToResponse()
	if row.DocTypeName != nil {
		resp.DocTypeName = *row.DocTypeName
	}
	if row.FolderName != nil {
		resp.FolderName = *row.FolderName
	}
	// Combine firstname + lastname for registrant name
	first := ""
	last := ""
	if row.RegistrantFirst != nil {
		first = *row.RegistrantFirst
	}
	if row.RegistrantLast != nil {
		last = *row.RegistrantLast
	}
	name := first
	if last != "" {
		if name != "" {
			name += " " + last
		} else {
			name = last
		}
	}
	resp.RegistrantName = name
	if row.RegistrantEmail != nil {
		resp.RegistrantEmail = *row.RegistrantEmail
	}
	if row.DepartmentName != nil {
		resp.DepartmentName = *row.DepartmentName
	}
	if row.SectorName != nil {
		resp.SectorName = *row.SectorName
	}
	return resp
}

const joinedDocSelect = `
	SELECT
		d.id, d.doc_no, d.doc_name, d.doc_path, d.type, d.doc_type_id, d.folder_id,
		d.registrant_id, d.status, d.version_number, d.description, d.send_to_director,
		d.created_at, d.updated_at,
		dt.type_name AS doc_type_name,
		f.folder_name,
		u.firstname AS registrant_first,
		u.lastname  AS registrant_last,
		u.email     AS registrant_email,
		dep.dept_name AS department_name,
		sec.name AS sector_name
	FROM docs d
	LEFT JOIN doc_types dt ON dt.id = d.doc_type_id
	LEFT JOIN folders   f  ON f.id  = d.folder_id
	LEFT JOIN users     u  ON u.id  = d.registrant_id
	LEFT JOIN departments dep ON dep.id = u.department_id
	LEFT JOIN sectors     sec ON sec.id = u.sector_id
`

func scanJoinedDoc(row pgx.Row) (*joinedDocScan, error) {
	var s joinedDocScan
	var desc sql.NullString
	var docPath sql.NullString
	var docType sql.NullString

	err := row.Scan(
		&s.ID, &s.DocNo, &s.DocName, &docPath, &docType, &s.DocTypeID, &s.FolderID,
		&s.RegistrantID, &s.Status, &s.VersionNumber, &desc, &s.SendToDirector,
		&s.CreatedAt, &s.UpdatedAt,
		&s.DocTypeName, &s.FolderName,
		&s.RegistrantFirst, &s.RegistrantLast, &s.RegistrantEmail,
		&s.DepartmentName, &s.SectorName,
	)
	if desc.Valid {
		s.Description = &desc.String
	}
	if docPath.Valid {
		s.DocPath = docPath.String
	}
	if docType.Valid {
		s.Type = docType.String
	}
	return &s, err
}

func (r *postgresRepository) FindAll(ctx context.Context, userID string) ([]domain.Document, error) {
	query := joinedDocSelect + ` ORDER BY d.created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.Document
	for rows.Next() {
		var s joinedDocScan
		var desc sql.NullString
		var docPath sql.NullString
		var docType sql.NullString
		if err := rows.Scan(
			&s.ID, &s.DocNo, &s.DocName, &docPath, &docType, &s.DocTypeID, &s.FolderID,
			&s.RegistrantID, &s.Status, &s.VersionNumber, &desc, &s.SendToDirector,
			&s.CreatedAt, &s.UpdatedAt,
			&s.DocTypeName, &s.FolderName,
			&s.RegistrantFirst, &s.RegistrantLast, &s.RegistrantEmail,
			&s.DepartmentName, &s.SectorName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		if desc.Valid {
			s.Description = &desc.String
		}
		if docPath.Valid {
			s.DocPath = docPath.String
		}
		if docType.Valid {
			s.Type = docType.String
		}
		documents = append(documents, s.Document)
	}

	return documents, nil
}

// FindAllJoined returns full DocumentResponse with joined fields
func (r *postgresRepository) FindAllJoined(ctx context.Context) ([]domain.DocumentResponse, error) {
	query := joinedDocSelect + ` ORDER BY d.created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer rows.Close()

	var responses []domain.DocumentResponse
	for rows.Next() {
		var s joinedDocScan
		var desc sql.NullString
		var docPath sql.NullString
		var docType sql.NullString
		if err := rows.Scan(
			&s.ID, &s.DocNo, &s.DocName, &docPath, &docType, &s.DocTypeID, &s.FolderID,
			&s.RegistrantID, &s.Status, &s.VersionNumber, &desc, &s.SendToDirector,
			&s.CreatedAt, &s.UpdatedAt,
			&s.DocTypeName, &s.FolderName,
			&s.RegistrantFirst, &s.RegistrantLast, &s.RegistrantEmail,
			&s.DepartmentName, &s.SectorName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		if desc.Valid {
			s.Description = &desc.String
		}
		if docPath.Valid {
			s.DocPath = docPath.String
		}
		if docType.Valid {
			s.Type = docType.String
		}
		responses = append(responses, buildDocResponse(&s))
	}

	return responses, nil
}

// FindByIDJoined returns a DocumentResponse with all joined fields for a single doc
func (r *postgresRepository) FindByIDJoined(ctx context.Context, id uuid.UUID) (*domain.DocumentResponse, error) {
	query := joinedDocSelect + ` WHERE d.id = $1`

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

// FindByFolderIDJoined returns DocumentResponse list with joined fields for a folder
func (r *postgresRepository) FindByFolderIDJoined(ctx context.Context, folderID uuid.UUID) ([]domain.DocumentResponse, error) {
	query := joinedDocSelect + ` WHERE d.folder_id = $1 ORDER BY d.created_at DESC`

	rows, err := r.pool.Query(ctx, query, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents by folder: %w", err)
	}
	defer rows.Close()

	var responses []domain.DocumentResponse
	for rows.Next() {
		var s joinedDocScan
		var desc sql.NullString
		var docPath sql.NullString
		var docType sql.NullString
		if err := rows.Scan(
			&s.ID, &s.DocNo, &s.DocName, &docPath, &docType, &s.DocTypeID, &s.FolderID,
			&s.RegistrantID, &s.Status, &s.VersionNumber, &desc, &s.SendToDirector,
			&s.CreatedAt, &s.UpdatedAt,
			&s.DocTypeName, &s.FolderName,
			&s.RegistrantFirst, &s.RegistrantLast, &s.RegistrantEmail,
			&s.DepartmentName, &s.SectorName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		if desc.Valid {
			s.Description = &desc.String
		}
		if docPath.Valid {
			s.DocPath = docPath.String
		}
		if docType.Valid {
			s.Type = docType.String
		}
		responses = append(responses, buildDocResponse(&s))
	}

	return responses, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	query := `
		SELECT id, doc_no, doc_name, doc_path, type, doc_type_id, folder_id,
		       registrant_id, status, version_number, description, send_to_director, created_at, updated_at
		FROM docs
		WHERE id = $1
	`

	var doc domain.Document
	var desc sql.NullString
	var docPath sql.NullString
	var docType sql.NullString
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.DocNo, &doc.DocName, &docPath, &docType, &doc.DocTypeID, &doc.FolderID,
		&doc.RegistrantID, &doc.Status, &doc.VersionNumber, &desc, &doc.SendToDirector,
		&doc.CreatedAt, &doc.UpdatedAt,
	)
	if desc.Valid {
		doc.Description = &desc.String
	}
	if docPath.Valid {
		doc.DocPath = docPath.String
	}
	if docType.Valid {
		doc.Type = docType.String
	}
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	return &doc, nil
}

func (r *postgresRepository) FindByFolderID(ctx context.Context, folderID uuid.UUID) ([]domain.Document, error) {
	query := `
		SELECT id, doc_no, doc_name, doc_path, type, doc_type_id, folder_id,
		       registrant_id, status, version_number, description, send_to_director, created_at, updated_at
		FROM docs
		WHERE folder_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents by folder: %w", err)
	}
	defer rows.Close()

	var documents []domain.Document
	for rows.Next() {
		var doc domain.Document
		var desc sql.NullString
		var docPath sql.NullString
		var docType sql.NullString
		if err := rows.Scan(
			&doc.ID, &doc.DocNo, &doc.DocName, &docPath, &docType, &doc.DocTypeID, &doc.FolderID,
			&doc.RegistrantID, &doc.Status, &doc.VersionNumber, &desc, &doc.SendToDirector,
			&doc.CreatedAt, &doc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		if desc.Valid {
			doc.Description = &desc.String
		}
		if docPath.Valid {
			doc.DocPath = docPath.String
		}
		if docType.Valid {
			doc.Type = docType.String
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByDocNo(ctx context.Context, docNo string) (*domain.Document, error) {
	query := `
		SELECT id, doc_no, doc_name, doc_path, type, doc_type_id, folder_id,
		       registrant_id, status, version_number, description, send_to_director, created_at, updated_at
		FROM docs
		WHERE doc_no = $1
	`

	var doc domain.Document
	var desc sql.NullString
	var docPath sql.NullString
	var docType sql.NullString
	err := r.pool.QueryRow(ctx, query, docNo).Scan(
		&doc.ID, &doc.DocNo, &doc.DocName, &docPath, &docType, &doc.DocTypeID, &doc.FolderID,
		&doc.RegistrantID, &doc.Status, &doc.VersionNumber, &desc, &doc.SendToDirector,
		&doc.CreatedAt, &doc.UpdatedAt,
	)
	if desc.Valid {
		doc.Description = &desc.String
	}
	if docPath.Valid {
		doc.DocPath = docPath.String
	}
	if docType.Valid {
		doc.Type = docType.String
	}
	if err == pgx.ErrNoRows {
		return nil, nil // Not found is not an error for checking existence
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find document by doc_no: %w", err)
	}

	return &doc, nil
}

func (r *postgresRepository) Create(ctx context.Context, doc *domain.Document) error {
	query := `
		INSERT INTO docs (doc_no, doc_name, doc_path, type, doc_type_id, folder_id, registrant_id,
		                  status, version_number, description, send_to_director, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query,
		doc.DocNo, doc.DocName, doc.DocPath, doc.Type, doc.DocTypeID, doc.FolderID, doc.RegistrantID,
		doc.Status, doc.VersionNumber, doc.Description, doc.SendToDirector, doc.CreatedAt, doc.UpdatedAt,
	).Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id uuid.UUID, doc *domain.Document) error {
	query := `
		UPDATE docs 
		SET doc_name = $1, doc_type_id = $2, folder_id = $3, description = $4,
		    send_to_director = $5, status = $6, updated_at = $7, version_number = $8
		WHERE id = $9
	`

	doc.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query,
		doc.DocName, doc.DocTypeID, doc.FolderID, doc.Description,
		doc.SendToDirector, doc.Status, doc.UpdatedAt, doc.VersionNumber, id,
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
	query := `DELETE FROM docs WHERE id = $1`

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
		INSERT INTO versions (doc_id, version_number, doc_path, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	version.CreatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, version.DocID, version.VersionNumber, version.DocPath, version.CreatedAt).
		Scan(&version.ID, &version.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetVersionsByDocID(ctx context.Context, docID uuid.UUID) ([]domain.Version, error) {
	query := `
		SELECT id, doc_id, version_number, doc_path, created_at
		FROM versions
		WHERE doc_id = $1
		ORDER BY version_number DESC
	`

	rows, err := r.pool.Query(ctx, query, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	defer rows.Close()

	var versions []domain.Version
	for rows.Next() {
		var v domain.Version
		if err := rows.Scan(&v.ID, &v.DocID, &v.VersionNumber, &v.DocPath, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}

// setRegistrantID is a helper to parse and store the caller's UUID into docs.registrant_id
func parseUserUUID(userID string) *uuid.UUID {
	if userID == "" {
		return nil
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil
	}
	return &id
}
