package upload

import (
	"context"
	"e-document-backend/internal/app/incomingdoc"
	"e-document-backend/internal/app/outgoingdoc"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Service interface {
	ProcessUploadComplete(ctx context.Context, params ProcessUploadParams) (*ProcessUploadResult, error)
	GetFolder(ctx context.Context, folderID int) (*domain.Folder, error)
	GetVersionWithDoc(ctx context.Context, versionID uuid.UUID) (*domain.Version, *domain.DocDetails, error)
	GetFolderVersions(ctx context.Context, folderID int) ([]*domain.Version, error)
}

type ProcessUploadParams struct {
	RelativePath   string
	ParentFolderID *uuid.UUID
	OwnerID        uuid.UUID
	FilePath       string
	FileSize       int64
	FileType       string
	UploadID       string
	TargetModule   string
	ExtraMetadata  map[string]string
}

type ProcessUploadResult struct {
	Document *domain.DocDetails `json:"document"`
	Version  *domain.Version    `json:"version"`
	Folders  []*domain.Folder   `json:"folders"`
}

type service struct {
	repo            Repository
	incomingService incomingdoc.Service
	outgoingService outgoingdoc.Service
	userService     user.Service
}

func NewService(repo Repository, incomingService incomingdoc.Service, outgoingService outgoingdoc.Service, userService user.Service) Service {
	return &service{
		repo:            repo,
		incomingService: incomingService,
		outgoingService: outgoingService,
		userService:     userService,
	}
}

func (s *service) ProcessUploadComplete(ctx context.Context, params ProcessUploadParams) (*ProcessUploadResult, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg("failed to rollback transaction")
			}
		}
	}()

	result := &ProcessUploadResult{
		Folders: make([]*domain.Folder, 0),
	}

	pathParts := parsePath(params.RelativePath)
	if len(pathParts) == 0 {
		err = fmt.Errorf("invalid relative path: %s", params.RelativePath)
		return nil, err
	}

	fileName := pathParts[len(pathParts)-1]
	folderParts := pathParts[:len(pathParts)-1]

	var currentParentID *uuid.UUID = params.ParentFolderID
	var currentPath string

	for _, folderName := range folderParts {
		if currentPath == "" {
			currentPath = folderName
		} else {
			currentPath = currentPath + "/" + folderName
		}

		folder, findErr := s.repo.FindFolderByNameAndParent(ctx, tx, folderName, currentParentID, params.OwnerID.String())
		if findErr != nil {
			err = findErr
			return nil, err
		}

		if folder == nil {
			folder = &domain.Folder{
				FolderName:     folderName,
				FolderPath:     currentPath,
				UserID:         params.OwnerID,
				ParentFolderID: currentParentID,
			}

			if createErr := s.repo.CreateFolder(ctx, tx, folder); createErr != nil {
				err = createErr
				return nil, err
			}
		}

		result.Folders = append(result.Folders, folder)
		currentParentID = &folder.ID
	}

	titleWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	var existingDoc *domain.DocDetails
	// Outgoing and incoming docs always get a fresh doc_no —
	// each send/receive is an independent document, not a new version of a previous one.
	if params.TargetModule != "outgoing" && params.TargetModule != "incoming" {
		existingDoc, err = s.repo.FindDocumentByNameAndFolder(ctx, tx, titleWithoutExt, currentParentID)
		if err != nil {
			return nil, err
		}
	}

	var doc *domain.DocDetails
	var newVersionNumber int

	if existingDoc == nil {
		// Outgoing docs use the user-entered doc_no (required, must be unique).
		// Everything else falls back to an auto-generated number.
		docNo := util.GenerateLALDocNumber()
		if params.TargetModule == "outgoing" {
			provided := strings.TrimSpace(params.ExtraMetadata["doc_no"])
			if provided == "" {
				err = fmt.Errorf("doc_no is required for outgoing documents")
				return nil, err
			}
			existing, dErr := s.repo.FindDocumentByDocNo(ctx, tx, provided)
			if dErr != nil {
				err = dErr
				return nil, err
			}
			if existing != nil {
				err = fmt.Errorf("doc_no %q already exists", provided)
				return nil, err
			}
			docNo = provided
		}
		doc = &domain.DocDetails{
			DocNo:         docNo,
			DocName:       titleWithoutExt,
			UserID:        &params.OwnerID,
			Status:        domain.StatusNone,
			VersionNumber: 1,
		}

		if createErr := s.repo.CreateDocument(ctx, tx, doc); createErr != nil {
			err = createErr
			return nil, err
		}
		newVersionNumber = 1
	} else {
		latestVersion, vErr := s.repo.GetLatestVersionByDocumentID(ctx, tx, existingDoc.ID)
		if vErr != nil {
			err = vErr
			return nil, err
		}
		newVersionNumber = latestVersion + 1

		if updateErr := s.repo.UpdateDocumentVersion(ctx, tx, existingDoc.ID, newVersionNumber); updateErr != nil {
			err = updateErr
			return nil, err
		}

		doc = existingDoc
		doc.VersionNumber = newVersionNumber
	}

	result.Document = doc

	// Extract file type from the actual filename (e.g., "pdf" from "report.pdf").
	// params.FilePath is the TUS/S3 object key (a hash) and has no extension.
	// fileName was parsed from RelativePath which contains the real filename.
	fileType := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")

	version := &domain.Version{
		DocDetailsID:  doc.ID,
		FolderID:      currentParentID,
		VersionNumber: newVersionNumber,
		DocPath:       params.FilePath,
		FileType:      fileType,
	}
	if createErr := s.repo.CreateVersion(ctx, tx, version); createErr != nil {
		err = createErr
		return nil, err
	}
	result.Version = version

	if commitErr := tx.Commit(ctx); commitErr != nil {
		err = commitErr
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if params.TargetModule != "" {
		go s.postProcessModule(ctx, params, result.Document.ID)
	}

	return result, nil
}

func (s *service) postProcessModule(ctx context.Context, params ProcessUploadParams, docID uuid.UUID) {
	bgCtx := context.Background()
	switch params.TargetModule {
	case "incoming":
		s.handleIncomingModule(bgCtx, params, docID, uuid.Nil)
	case "outgoing":
		s.handleOutgoingModule(bgCtx, params, docID)
	}
}

// handleIncomingModule handles a standalone "incoming" upload (legacy path):
// creates incoming docs directly for each department, no sequential route.
func (s *service) handleIncomingModule(ctx context.Context, params ProcessUploadParams, docID uuid.UUID, outgoingDocID uuid.UUID) {
	deptUUIDs := parseDeptIDs(params.ExtraMetadata["receiver_ids"])
	if len(deptUUIDs) == 0 {
		log.Warn().Str("upload_id", params.UploadID).Msg("Incoming module requested but no receiver_ids provided")
		return
	}

	remark := params.ExtraMetadata["description"]
	if err := s.incomingService.CreateDirectIncomingDocs(ctx, docID, outgoingDocID, &params.OwnerID, deptUUIDs, remark); err != nil {
		log.Error().Err(err).Str("doc_id", docID.String()).Msg("Failed to create incoming docs")
	}
}

// handleOutgoingModule builds the ordered recipient route and creates the
// outgoing doc in the "pending" state. No incoming docs are created here — the
// owner department head must first approve the outgoing doc, which then kicks off
// the sequential recipient flow.
// Route order: selected departments (in order) → Director Office (optional).
func (s *service) handleOutgoingModule(ctx context.Context, params ProcessUploadParams, docID uuid.UUID) {
	// Resolve owner's department (used to exclude it from the recipient route).
	// If the owner has no department (e.g. a system admin), ownerDept stays
	// uuid.Nil and BuildRoute simply excludes nothing.
	userRes, err := s.userService.GetUserByID(ctx, params.OwnerID.String())
	var ownerDept uuid.UUID
	var ownerDeptID *uuid.UUID
	if err == nil && userRes.DepartmentID != nil {
		ownerDept = *userRes.DepartmentID
		ownerDeptID = userRes.DepartmentID
	}

	selected := parseDeptIDs(params.ExtraMetadata["receiver_ids"])
	route := outgoingdoc.BuildRoute(ownerDept, selected, nil)

	if _, err := s.outgoingService.CreateOutgoingDocWithRoute(ctx, docID, &params.OwnerID, ownerDeptID, route); err != nil {
		log.Error().Err(err).Str("doc_id", docID.String()).Msg("Failed to create outgoing doc with route")
	}
}

// parseDeptIDs parses a comma-separated list of department UUIDs, preserving order.
func parseDeptIDs(csv string) []uuid.UUID {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		if id, err := uuid.Parse(strings.TrimSpace(p)); err == nil {
			out = append(out, id)
		}
	}
	return out
}

func parsePath(path string) []string {
	normalized := strings.ReplaceAll(path, "\\", "/")
	normalized = strings.Trim(normalized, "/")
	if normalized == "" {
		return []string{}
	}
	parts := strings.Split(normalized, "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func (s *service) GetFolder(ctx context.Context, folderID int) (*domain.Folder, error) {
	return s.repo.GetFolderByID(ctx, folderID)
}

func (s *service) GetVersionWithDoc(ctx context.Context, versionID uuid.UUID) (*domain.Version, *domain.DocDetails, error) {
	return s.repo.GetVersionWithDoc(ctx, versionID)
}

func (s *service) GetFolderVersions(ctx context.Context, folderID int) ([]*domain.Version, error) {
	return s.repo.GetVersionsByFolderID(ctx, folderID)
}
