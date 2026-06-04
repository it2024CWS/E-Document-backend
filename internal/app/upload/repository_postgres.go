package upload

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

func (r *postgresRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *postgresRepository) FindFolderByNameAndParent(ctx context.Context, tx pgx.Tx, name string, parentID *uuid.UUID, userID string) (*domain.Folder, error) {
	var query string
	var args []interface{}
	query = `
		SELECT id, folder_name, folder_path, user_id, parent_folder_id, created_at
		FROM folders
		WHERE folder_name = $1
		  AND user_id = $2
		  AND deleted_at IS NULL`
	if parentID == nil {
		query += ` AND parent_folder_id IS NULL`
		args = []interface{}{name, userID}
	} else {
		query += ` AND parent_folder_id = $3`
		args = []interface{}{name, userID, parentID}
	}
	query += ` LIMIT 1`

	var folder domain.Folder
	err := tx.QueryRow(ctx, query, args...).Scan(
		&folder.ID,
		&folder.FolderName,
		&folder.FolderPath,
		&folder.UserID,
		&folder.ParentFolderID,
		&folder.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find folder: %w", err)
	}
	return &folder, nil
}

func (r *postgresRepository) CreateFolder(ctx context.Context, tx pgx.Tx, folder *domain.Folder) error {
	query := `
		INSERT INTO folders (folder_name, folder_path, user_id, parent_folder_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	folder.CreatedAt = time.Now().UTC()
	return tx.QueryRow(ctx, query,
		folder.FolderName,
		folder.FolderPath,
		folder.UserID,
		folder.ParentFolderID,
		folder.CreatedAt,
	).Scan(&folder.ID)
}

func (r *postgresRepository) GetFolderByID(ctx context.Context, folderID int) (*domain.Folder, error) {
	query := `
		SELECT id, folder_name, folder_path, user_id, parent_folder_id, created_at
		FROM folders
		WHERE id = $1 AND deleted_at IS NULL
	`
	var folder domain.Folder
	err := r.pool.QueryRow(ctx, query, folderID).Scan(
		&folder.ID,
		&folder.FolderName,
		&folder.FolderPath,
		&folder.UserID,
		&folder.ParentFolderID,
		&folder.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}
	return &folder, nil
}

func (r *postgresRepository) FindDocumentByNameAndFolder(ctx context.Context, tx pgx.Tx, docName string, folderID *uuid.UUID) (*domain.DocDetails, error) {
	query := `
		SELECT d.id, d.doc_no, d.doc_name, d.description, d.version_number, d.status, d.doc_type_id, d.user_id, d.created_at, d.updated_at
		FROM doc_details d
		JOIN versions v ON v.doc_details_id = d.id AND v.version_number = d.version_number
		WHERE d.doc_name = $1 AND d.deleted_at IS NULL
	`
	var args []interface{}
	args = append(args, docName)

	if folderID == nil {
		query += ` AND v.folder_id IS NULL`
	} else {
		query += ` AND v.folder_id = $2`
		args = append(args, folderID)
	}
	query += ` LIMIT 1`

	var doc domain.DocDetails
	var desc *string
	err := tx.QueryRow(ctx, query, args...).Scan(
		&doc.ID, &doc.DocNo, &doc.DocName, &desc, &doc.VersionNumber, &doc.Status,
		&doc.DocTypeID, &doc.UserID, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find document by name: %w", err)
	}
	doc.Description = desc
	return &doc, nil
}

func (r *postgresRepository) CreateDocument(ctx context.Context, tx pgx.Tx, doc *domain.DocDetails) error {
	query := `
		INSERT INTO doc_details (doc_no, doc_name, version_number, status, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	return tx.QueryRow(ctx, query,
		doc.DocNo,
		doc.DocName,
		doc.VersionNumber,
		doc.Status,
		doc.UserID,
		doc.CreatedAt,
		doc.UpdatedAt,
	).Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
}

func (r *postgresRepository) CreateVersion(ctx context.Context, tx pgx.Tx, version *domain.Version) error {
	query := `
		INSERT INTO versions (doc_details_id, folder_id, version_number, doc_path, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	version.CreatedAt = time.Now().UTC()
	return tx.QueryRow(ctx, query,
		version.DocDetailsID,
		version.FolderID,
		version.VersionNumber,
		version.DocPath,
		version.CreatedAt,
	).Scan(&version.ID, &version.CreatedAt)
}

func (r *postgresRepository) GetLatestVersionByDocumentID(ctx context.Context, tx pgx.Tx, documentID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(version_number), 0) FROM versions WHERE doc_details_id = $1`
	var version int
	if err := tx.QueryRow(ctx, query, documentID).Scan(&version); err != nil {
		return 0, fmt.Errorf("failed to get latest version: %w", err)
	}
	return version, nil
}

func (r *postgresRepository) UpdateDocumentVersion(ctx context.Context, tx pgx.Tx, docID uuid.UUID, newVersionNumber int) error {
	query := `UPDATE doc_details SET version_number = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
	if _, err := tx.Exec(ctx, query, newVersionNumber, time.Now().UTC(), docID); err != nil {
		return fmt.Errorf("failed to update document version: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetVersionWithDoc(ctx context.Context, versionID uuid.UUID) (*domain.Version, *domain.DocDetails, error) {
	query := `
		SELECT v.id, v.doc_details_id, v.folder_id, v.version_number, v.doc_path, v.created_at,
		       d.id, d.doc_no, d.doc_name, d.description, d.version_number, d.status, d.doc_type_id, d.user_id, d.created_at, d.updated_at
		FROM versions v
		JOIN doc_details d ON v.doc_details_id = d.id
		WHERE v.id = $1 AND d.deleted_at IS NULL
	`

	var v domain.Version
	var d domain.DocDetails
	var desc *string

	err := r.pool.QueryRow(ctx, query, versionID).Scan(
		&v.ID, &v.DocDetailsID, &v.FolderID, &v.VersionNumber, &v.DocPath, &v.CreatedAt,
		&d.ID, &d.DocNo, &d.DocName, &desc, &d.VersionNumber, &d.Status, &d.DocTypeID, &d.UserID, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, fmt.Errorf("version not found")
		}
		return nil, nil, fmt.Errorf("failed to get version: %w", err)
	}
	d.Description = desc

	return &v, &d, nil
}

func (r *postgresRepository) GetVersionsByFolderID(ctx context.Context, folderID int) ([]*domain.Version, error) {
	query := `
		WITH RECURSIVE folder_tree AS (
			SELECT id FROM folders WHERE id = $1 AND deleted_at IS NULL
			UNION ALL
			SELECT f.id FROM folders f
			INNER JOIN folder_tree ft ON f.parent_folder_id = ft.id
			WHERE f.deleted_at IS NULL
		)
		SELECT DISTINCT
			v.id, v.doc_details_id, v.folder_id, v.version_number, v.doc_path, v.created_at
		FROM versions v
		INNER JOIN doc_details d ON d.id = v.doc_details_id
		INNER JOIN folder_tree ft ON v.folder_id = ft.id
		WHERE d.deleted_at IS NULL
		ORDER BY v.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions by folder: %w", err)
	}
	defer rows.Close()

	var versions []*domain.Version
	for rows.Next() {
		var v domain.Version
		if err := rows.Scan(
			&v.ID, &v.DocDetailsID, &v.FolderID, &v.VersionNumber, &v.DocPath, &v.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, &v)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating versions: %w", err)
	}
	return versions, nil
}

