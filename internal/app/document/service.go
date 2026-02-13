package document

import (
	"context"
	"e-document-backend/internal/app/incomingdoc"
	"e-document-backend/internal/app/outgoingdoc"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
	"time"
)

// Service defines the interface for document business logic
type Service interface {
	GetAllDocuments(ctx context.Context, userID string) ([]domain.DocumentResponse, error)
	GetDocumentByID(ctx context.Context, id int) (*domain.DocumentResponse, error)
	GetDocumentsByFolder(ctx context.Context, folderID int) ([]domain.DocumentResponse, error)
	CreateDocument(ctx context.Context, userID string, req domain.CreateDocumentRequest) (*domain.DocumentResponse, error)
	UpdateDocument(ctx context.Context, id int, req domain.UpdateDocumentRequest) (*domain.DocumentResponse, error)
	DeleteDocument(ctx context.Context, id int) error
	GetDocumentVersions(ctx context.Context, id int) ([]domain.VersionResponse, error)
	SendDocument(ctx context.Context, userID string, id int, req domain.SendDocumentRequest) error
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

// GetAllDocuments retrieves all documents
func (s *service) GetAllDocuments(ctx context.Context, userID string) ([]domain.DocumentResponse, error) {
	docs, err := s.repo.FindAll(ctx, userID)
	if err != nil {
		return nil, util.NewDatabaseError("get all documents", err)
	}

	responses := make([]domain.DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, nil
}

// GetDocumentByID retrieves a document by ID
func (s *service) GetDocumentByID(ctx context.Context, id int) (*domain.DocumentResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Document", fmt.Sprintf("%d", id))
	}

	response := doc.ToResponse()
	return &response, nil
}

// GetDocumentsByFolder retrieves documents by folder ID
func (s *service) GetDocumentsByFolder(ctx context.Context, folderID int) ([]domain.DocumentResponse, error) {
	docs, err := s.repo.FindByFolderID(ctx, folderID)
	if err != nil {
		return nil, util.NewDatabaseError("get documents by folder", err)
	}

	responses := make([]domain.DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
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

	// Create document object
	doc := &domain.Document{
		DocNo:          docNo,
		DocName:        req.DocName,
		DocPath:        req.DocPath, // Set from request
		DocTypeID:      req.DocTypeID,
		FolderID:       req.FolderID,
		Description:    req.Description,
		SendToDirector: req.SendToDirector,
		Status:         domain.StatusNone,
		VersionNumber:  1,
	}

	// Save to database
	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, util.NewDatabaseError("create document", err)
	}

	response := doc.ToResponse()
	return &response, nil
}

// UpdateDocument updates a document
func (s *service) UpdateDocument(ctx context.Context, id int, req domain.UpdateDocumentRequest) (*domain.DocumentResponse, error) {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	// Check if document exists
	existingDoc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Document", fmt.Sprintf("%d", id))
	}

	// Handle Versioning if File changes
	if req.DocPath != "" && req.DocPath != existingDoc.DocPath {
		// Archive current version
		version := &domain.Version{
			DocID:         existingDoc.ID,
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
		existingDoc.Description = req.Description
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
func (s *service) GetDocumentVersions(ctx context.Context, id int) ([]domain.VersionResponse, error) {
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
func (s *service) DeleteDocument(ctx context.Context, id int) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return util.NewNotFoundError("Document", fmt.Sprintf("%d", id))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return util.NewDatabaseError("delete document", err)
	}

	return nil
}

// SendDocument sends a document to a receiver (internal routing)
func (s *service) SendDocument(ctx context.Context, userID string, id int, req domain.SendDocumentRequest) error {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return err
	}

	// Check if document exists
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return util.NewNotFoundError("Document", fmt.Sprintf("%d", id))
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
