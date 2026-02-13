package document

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"
	"time"

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

func (r *postgresRepository) FindAll(ctx context.Context, userID string) ([]domain.Document, error) {
	query := `
		SELECT id, doc_no, doc_name, doc_path, type, doc_type_id, folder_id, status, version_number, description, send_to_director, created_at, updated_at
		FROM docs
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer rows.Close()

	var documents []domain.Document
	for rows.Next() {
		var doc domain.Document
		if err := rows.Scan(&doc.ID, &doc.DocNo, &doc.DocName, &doc.DocPath, &doc.Type, &doc.DocTypeID, &doc.FolderID, &doc.Status, &doc.VersionNumber, &doc.Description, &doc.SendToDirector, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id int) (*domain.Document, error) {
	query := `
		SELECT id, doc_no, doc_name, doc_path, type, doc_type_id, folder_id, status, version_number, description, send_to_director, created_at, updated_at
		FROM docs
		WHERE id = $1
	`

	var doc domain.Document
	err := r.pool.QueryRow(ctx, query, id).Scan(&doc.ID, &doc.DocNo, &doc.DocName, &doc.DocPath, &doc.Type, &doc.DocTypeID, &doc.FolderID, &doc.Status, &doc.VersionNumber, &doc.Description, &doc.SendToDirector, &doc.CreatedAt, &doc.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	return &doc, nil
}

func (r *postgresRepository) FindByFolderID(ctx context.Context, folderID int) ([]domain.Document, error) {
	query := `
		SELECT id, doc_no, doc_name, doc_path, type, doc_type_id, folder_id, status, version_number, description, send_to_director, created_at, updated_at
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
		if err := rows.Scan(&doc.ID, &doc.DocNo, &doc.DocName, &doc.DocPath, &doc.Type, &doc.DocTypeID, &doc.FolderID, &doc.Status, &doc.VersionNumber, &doc.Description, &doc.SendToDirector, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (r *postgresRepository) FindByDocNo(ctx context.Context, docNo string) (*domain.Document, error) {
	query := `
		SELECT id, doc_no, doc_name, doc_path, type, doc_type_id, folder_id, status, version_number, description, send_to_director, created_at, updated_at
		FROM docs
		WHERE doc_no = $1
	`

	var doc domain.Document
	err := r.pool.QueryRow(ctx, query, docNo).Scan(&doc.ID, &doc.DocNo, &doc.DocName, &doc.DocPath, &doc.Type, &doc.DocTypeID, &doc.FolderID, &doc.Status, &doc.VersionNumber, &doc.Description, &doc.SendToDirector, &doc.CreatedAt, &doc.UpdatedAt)
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
		INSERT INTO docs (doc_no, doc_name, doc_path, type, doc_type_id, folder_id, status, version_number, description, send_to_director, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, doc.DocNo, doc.DocName, doc.DocPath, doc.Type, doc.DocTypeID, doc.FolderID, doc.Status, doc.VersionNumber, doc.Description, doc.SendToDirector, doc.CreatedAt, doc.UpdatedAt).
		Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id int, doc *domain.Document) error {
	// Update version_number as well
	query := `
		UPDATE docs 
		SET doc_name = $1, doc_type_id = $2, folder_id = $3, description = $4, send_to_director = $5, status = $6, updated_at = $7, version_number = $8
		WHERE id = $9
	`

	doc.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query, doc.DocName, doc.DocTypeID, doc.FolderID, doc.Description, doc.SendToDirector, doc.Status, doc.UpdatedAt, doc.VersionNumber, id)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id int) error {
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

func (r *postgresRepository) GetVersionsByDocID(ctx context.Context, docID int) ([]domain.Version, error) {
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
