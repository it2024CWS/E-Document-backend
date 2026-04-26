package document

import (
	"context"
	"e-document-backend/internal/app/incomingdoc"
	"e-document-backend/internal/app/outgoingdoc"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"time"

	"github.com/google/uuid"
)

// Service defines the interface for document business logic
type Service interface {
	GetAllDocuments(ctx context.Context, userID string) ([]domain.DocumentResponse, error)
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*domain.DocumentDetailResponse, error)
	GetDocumentsByFolder(ctx context.Context, folderID uuid.UUID) ([]domain.DocumentResponse, error)
	CreateDocument(ctx context.Context, userID string, req domain.CreateDocumentRequest) (*domain.DocumentResponse, error)
	UpdateDocument(ctx context.Context, id uuid.UUID, req domain.UpdateDocumentRequest) (*domain.DocumentResponse, error)
	DeleteDocument(ctx context.Context, id uuid.UUID) error
	GetDocumentVersions(ctx context.Context, id uuid.UUID) ([]domain.VersionResponse, error)
	SendDocument(ctx context.Context, userID string, id uuid.UUID, req domain.SendDocumentRequest) error
}

type service struct {
	repo         Repository
	incomingRepo incomingdoc.Repository
	outgoingRepo outgoingdoc.Repository
}

// NewService creates a new document service
func NewService(repo Repository, incomingRepo incomingdoc.Repository, outgoingRepo outgoingdoc.Repository) Service {
	return &service{
		repo:         repo,
		incomingRepo: incomingRepo,
		outgoingRepo: outgoingRepo,
	}
}

// GetAllDocuments retrieves all documents with joined fields
func (s *service) GetAllDocuments(ctx context.Context, userID string) ([]domain.DocumentResponse, error) {
	responses, err := s.repo.FindAllJoined(ctx)
	if err != nil {
		return nil, util.NewDatabaseError("get all documents", err)
	}
	return responses, nil
}

// GetDocumentByID retrieves a document by ID with joined fields and its versions
func (s *service) GetDocumentByID(ctx context.Context, id uuid.UUID) (*domain.DocumentDetailResponse, error) {
	doc, err := s.repo.FindByIDJoined(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Document", id.String())
	}

	// Attach versions
	versions, err := s.repo.GetVersionsByDocID(ctx, id)
	if err == nil {
		responses := make([]domain.VersionResponse, len(versions))
		for i, v := range versions {
			responses[i] = v.ToResponse()
		}
		doc.Versions = responses
	}

	return doc, nil
}

// GetDocumentsByFolder retrieves documents by folder ID with joined fields
func (s *service) GetDocumentsByFolder(ctx context.Context, folderID uuid.UUID) ([]domain.DocumentResponse, error) {
	responses, err := s.repo.FindByFolderIDJoined(ctx, folderID)
	if err != nil {
		return nil, util.NewDatabaseError("get documents by folder", err)
	}
	if responses == nil {
		responses = []domain.DocumentResponse{}
	}
	return responses, nil
}

// CreateDocument creates a new document
func (s *service) CreateDocument(ctx context.Context, userID string, req domain.CreateDocumentRequest) (*domain.DocumentResponse, error) {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	// Generate document number using LAL format
	docNo := util.GenerateLALDocNumber()

	// Parse caller UUID for registrant_id
	registrantID := parseUserUUID(userID)

	// Create document object
	doc := &domain.Document{
		DocNo:         docNo,
		DocName:       req.DocName,
		DocPath:       req.DocPath,
		DocTypeID:     req.DocTypeID,
		FolderID:      req.FolderID,
		RegistrantID:  registrantID,
		Status:        domain.StatusNone,
		VersionNumber: 1,
	}
	if req.Description != "" {
		doc.Description = &req.Description
	}

	// Save to database
	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, util.NewDatabaseError("create document", err)
	}

	// Return joined response (unwrap to DocumentResponse since it's a new doc with no versions)
	detail, err := s.repo.FindByIDJoined(ctx, doc.ID)
	if err != nil {
		return nil, util.NewDatabaseError("fetch created document", err)
	}
	resp := detail.DocumentResponse
	return &resp, nil
}

// UpdateDocument updates a document
func (s *service) UpdateDocument(ctx context.Context, id uuid.UUID, req domain.UpdateDocumentRequest) (*domain.DocumentResponse, error) {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	// Check if document exists
	existingDoc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Document", id.String())
	}

	// Handle Versioning if File changes
	if req.DocPath != "" && req.DocPath != existingDoc.DocPath {
		// Archive current version
		version := &domain.Version{
			DocID:         existingDoc.ID,
			FolderID:      existingDoc.FolderID,
			VersionNumber: existingDoc.VersionNumber,
			DocPath:       existingDoc.DocPath,
		}
		if err := s.repo.CreateVersion(ctx, version); err != nil {
			return nil, util.NewDatabaseError("create version", err)
		}

		// Increment version number and update path
		existingDoc.VersionNumber++
		existingDoc.DocPath = req.DocPath
	}

	// Update fields if provided
	if req.DocName != "" {
		existingDoc.DocName = req.DocName
	}
	if req.DocTypeID != nil {
		existingDoc.DocTypeID = req.DocTypeID
	}
	if req.FolderID != nil {
		existingDoc.FolderID = req.FolderID
	}
	if req.Description != "" {
		existingDoc.Description = &req.Description
	} else {
		existingDoc.Description = nil
	}
	if req.SendToDirector != nil {
		existingDoc.SendToDirector = *req.SendToDirector
	}

	// Update in database
	if err := s.repo.Update(ctx, id, existingDoc); err != nil {
		return nil, util.NewDatabaseError("update document", err)
	}

	// Fetch updated document
	updatedDoc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewDatabaseError("get updated document", err)
	}

	response := updatedDoc.ToResponse()
	return &response, nil
}

// GetDocumentVersions retrieves versions of a document
func (s *service) GetDocumentVersions(ctx context.Context, id uuid.UUID) ([]domain.VersionResponse, error) {
	versions, err := s.repo.GetVersionsByDocID(ctx, id)
	if err != nil {
		return nil, util.NewDatabaseError("get document versions", err)
	}

	responses := make([]domain.VersionResponse, len(versions))
	for i, v := range versions {
		responses[i] = v.ToResponse()
	}

	return responses, nil
}

// DeleteDocument deletes a document
func (s *service) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return util.NewNotFoundError("Document", id.String())
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return util.NewDatabaseError("delete document", err)
	}

	return nil
}

// SendDocument sends a document to a receiver (internal routing)
func (s *service) SendDocument(ctx context.Context, userID string, id uuid.UUID, req domain.SendDocumentRequest) error {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return err
	}

	// Check if document exists
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return util.NewNotFoundError("Document", id.String())
	}

	// Parse User ID
	senderUUID, err := util.ParseUUID(userID)
	if err != nil {
		return util.NewInvalidInputError("user_id", "invalid UUID format")
	}

	// Parse Receiver ID
	receiverUUID, err := util.ParseUUID(req.ReceiverID)
	if err != nil {
		return util.NewInvalidInputError("receiver_id", "invalid UUID format")
	}

	// Create Outgoing Doc Record
	outgoing := &domain.OutgoingDoc{
		OutgoingNo: util.GenerateOutgoingNumber(),
		DocID:      doc.ID,
		UserID:     &senderUUID,
		CreatedAt:  time.Now(),
	}
	if err := s.outgoingRepo.Create(ctx, outgoing); err != nil {
		return util.NewDatabaseError("create outgoing document", err)
	}

	// Create Incoming Doc Record
	incoming := &domain.IncomingDoc{
		IncomingNo: util.GenerateIncomingNumber(),
		DocID:      doc.ID,
		SenderID:   &senderUUID,
		ReceiverID: &receiverUUID,
		Status:     domain.IncomingStatusPending,
		Remark:     req.Remark,
		CreatedAt:  time.Now(),
	}
	if err := s.incomingRepo.Create(ctx, incoming); err != nil {
		return util.NewDatabaseError("create incoming document", err)
	}

	// Update Document Status
	doc.Status = domain.StatusPending
	if err := s.repo.Update(ctx, id, doc); err != nil {
		return util.NewDatabaseError("update document status", err)
	}

	return nil
}
