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

type Service interface {
	GetAllDocuments(ctx context.Context, userID string) ([]domain.DocumentResponse, error)
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*domain.DocumentResponse, error)
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

func NewService(repo Repository, incomingRepo incomingdoc.Repository, outgoingRepo outgoingdoc.Repository) Service {
	return &service{
		repo:         repo,
		incomingRepo: incomingRepo,
		outgoingRepo: outgoingRepo,
	}
}

func (s *service) GetAllDocuments(ctx context.Context, userID string) ([]domain.DocumentResponse, error) {
	responses, err := s.repo.FindAllJoined(ctx)
	if err != nil {
		return nil, util.NewDatabaseError("get all documents", err)
	}
	return responses, nil
}

func (s *service) GetDocumentByID(ctx context.Context, id uuid.UUID) (*domain.DocumentResponse, error) {
	doc, err := s.repo.FindByIDJoined(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Document", id.String())
	}
	return doc, nil
}

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

func (s *service) CreateDocument(ctx context.Context, userID string, req domain.CreateDocumentRequest) (*domain.DocumentResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	docNo := util.GenerateLALDocNumber()

	var userUUID *uuid.UUID
	if userID != "" {
		id, err := uuid.Parse(userID)
		if err == nil {
			userUUID = &id
		}
	}

	doc := &domain.DocDetails{
		DocNo:         docNo,
		DocName:       req.DocName,
		DocTypeID:     req.DocTypeID,
		UserID:        userUUID,
		Status:        domain.StatusNone,
		VersionNumber: 1,
	}
	if req.Description != "" {
		doc.Description = &req.Description
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, util.NewDatabaseError("create document", err)
	}

	return s.repo.FindByIDJoined(ctx, doc.ID)
}

func (s *service) UpdateDocument(ctx context.Context, id uuid.UUID, req domain.UpdateDocumentRequest) (*domain.DocumentResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	existingDoc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Document", id.String())
	}

	if req.DocName != "" {
		existingDoc.DocName = req.DocName
	}
	if req.DocTypeID != nil {
		existingDoc.DocTypeID = req.DocTypeID
	}
	if req.Description != "" {
		existingDoc.Description = &req.Description
	} else {
		existingDoc.Description = nil
	}

	if err := s.repo.Update(ctx, id, existingDoc); err != nil {
		return nil, util.NewDatabaseError("update document", err)
	}

	updatedDoc, err := s.repo.FindByIDJoined(ctx, id)
	if err != nil {
		return nil, util.NewDatabaseError("get updated document", err)
	}

	return updatedDoc, nil
}

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

func (s *service) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return util.NewNotFoundError("Document", id.String())
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return util.NewDatabaseError("delete document", err)
	}

	return nil
}

func (s *service) SendDocument(ctx context.Context, userID string, id uuid.UUID, req domain.SendDocumentRequest) error {
	if err := util.ValidateStruct(&req); err != nil {
		return err
	}

	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return util.NewNotFoundError("Document", id.String())
	}

	senderUUID, err := uuid.Parse(userID)
	if err != nil {
		return util.NewInvalidInputError("user_id", "invalid UUID format")
	}

	outgoing := &domain.OutgoingDoc{
		OutgoingNo:   util.GenerateOutgoingNumber(),
		DocDetailsID: doc.ID,
		CreatedBy:    &senderUUID,
		CreatedAt:    time.Now(),
	}
	if err := s.outgoingRepo.Create(ctx, outgoing); err != nil {
		return util.NewDatabaseError("create outgoing document", err)
	}

	incoming := &domain.IncomingDoc{
		IncomingNo:   util.GenerateIncomingNumber(),
		DocDetailsID: doc.ID,
		CreatedBy:    &senderUUID,
		Status:       domain.IncomingStatusPending,
		Remark:       req.Remark,
	}
	// TODO: map receiver properly in new ERD (it doesn't have receiver_id). Probably we set it to something else or updated_by.
	// For now just pass it through creation.
	if err := s.incomingRepo.Create(ctx, incoming); err != nil {
		return util.NewDatabaseError("create incoming document", err)
	}

	doc.Status = domain.StatusPending
	if err := s.repo.Update(ctx, id, doc); err != nil {
		return util.NewDatabaseError("update document status", err)
	}

	return nil
}
