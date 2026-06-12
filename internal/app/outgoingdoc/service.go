package outgoingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Service interface {
	GetAllOutgoingDocs(ctx context.Context, page, limit int) ([]domain.OutgoingDocResponse, int, error)
	GetOutgoingDocByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDocResponse, error)
	GetOutgoingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.OutgoingDocResponse, int, error)
	CreateOutgoingDocWithParams(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, deptID *uuid.UUID) (uuid.UUID, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// buildStatusCounts computes status counts from a recipients slice
func buildStatusCounts(recipients []domain.RecipientInfo) domain.StatusCounts {
	counts := domain.StatusCounts{Total: len(recipients)}
	for _, r := range recipients {
		switch r.Status {
		case domain.IncomingStatusPending:
			counts.Pending++
		case domain.IncomingStatusReceived:
			counts.Received++
		case domain.IncomingStatusApproved:
			counts.Approved++
		case domain.IncomingStatusRejected:
			counts.Rejected++
		}
	}
	return counts
}

// enrichResponse fetches recipients for a doc and fills them into the response
func (s *service) enrichResponse(ctx context.Context, resp *domain.OutgoingDocResponse) {
	recipients, err := s.repo.FindRecipientsByOutgoingDocID(ctx, resp.ID)
	if err != nil {
		log.Error().Err(err).Str("outgoing_doc_id", resp.ID.String()).Msg("failed to fetch recipients")
		recipients = []domain.RecipientInfo{}
	}
	if recipients == nil {
		recipients = []domain.RecipientInfo{}
	}
	resp.Recipients = recipients
	resp.StatusCounts = buildStatusCounts(recipients)
}

func (s *service) GetAllOutgoingDocs(ctx context.Context, page, limit int) ([]domain.OutgoingDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get all outgoing documents", err)
	}

	responses := make([]domain.OutgoingDocResponse, len(docs))
	for i, doc := range docs {
		resp := doc.ToResponse()
		s.enrichResponse(ctx, &resp)
		responses[i] = resp
	}
	return responses, total, nil
}

func (s *service) GetOutgoingDocByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("OutgoingDoc", id.String())
	}

	resp := doc.ToResponse()
	s.enrichResponse(ctx, &resp)
	return &resp, nil
}

func (s *service) GetOutgoingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.OutgoingDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindByDepartmentID(ctx, deptID, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get outgoing documents by department", err)
	}

	responses := make([]domain.OutgoingDocResponse, len(docs))
	for i, doc := range docs {
		resp := doc.ToResponse()
		s.enrichResponse(ctx, &resp)
		responses[i] = resp
	}
	return responses, total, nil
}

func (s *service) CreateOutgoingDocWithParams(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, deptID *uuid.UUID) (uuid.UUID, error) {
	doc := &domain.OutgoingDoc{
		OutgoingNo:   util.GenerateOutgoingNumber(),
		DocDetailsID: docDetailsID,
		CreatedBy:    creatorID,
	}
	if err := s.repo.Create(ctx, doc); err != nil {
		return uuid.Nil, util.NewDatabaseError("create outgoing document", err)
	}
	return doc.ID, nil
}
