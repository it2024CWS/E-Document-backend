package folder

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

func (r *postgresRepository) FindAll(ctx context.Context) ([]domain.FolderResponse, error) {
	query := `
		SELECT id, folder_name, folder_path, user_id, parent_folder_id, created_at
		FROM folders
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find folders: %w", err)
	}
	defer rows.Close()

	var folders []domain.FolderResponse
	for rows.Next() {
		var f domain.FolderResponse
		if err := rows.Scan(&f.ID, &f.FolderName, &f.FolderPath, &f.UserID, &f.ParentFolderID, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}
		folders = append(folders, f)
	}

	return folders, nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.FolderResponse, error) {
	query := `
		SELECT id, folder_name, folder_path, user_id, parent_folder_id, created_at
		FROM folders
		WHERE id = $1 AND deleted_at IS NULL
	`
	var f domain.FolderResponse
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&f.ID, &f.FolderName, &f.FolderPath, &f.UserID, &f.ParentFolderID, &f.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("folder not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find folder: %w", err)
	}
	return &f, nil
}

func (r *postgresRepository) Create(ctx context.Context, folder *domain.Folder) error {
	query := `
		INSERT INTO folders (folder_name, folder_path, user_id, parent_folder_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	folder.CreatedAt = time.Now()
	err := r.pool.QueryRow(ctx, query,
		folder.FolderName, folder.FolderPath, folder.UserID, folder.ParentFolderID, folder.CreatedAt,
	).Scan(&folder.ID, &folder.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id uuid.UUID, req domain.CreateFolderRequest) (*domain.FolderResponse, error) {
	query := `
		UPDATE folders
		SET folder_name = $1, folder_path = $2, parent_folder_id = $3
		WHERE id = $4 AND deleted_at IS NULL
	`
	result, err := r.pool.Exec(ctx, query, req.FolderName, req.FolderPath, req.ParentFolderID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update folder: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("folder not found")
	}
	return r.FindByID(ctx, id)
}

func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `UPDATE folders SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("folder not found")
	}
	return nil
}

// parseUserUUID safely parses a UUID string
func parseUserUUID(userID string) uuid.UUID {
	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil
	}
	return id
}
