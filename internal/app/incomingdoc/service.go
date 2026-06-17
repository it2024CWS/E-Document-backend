package incomingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// RouteAdvancer exposes the route operations the incoming-doc flow needs to
// advance an outgoing document to the next department. Implemented by the
// outgoingdoc repository.
type RouteAdvancer interface {
	FindRouteByIncomingDocID(ctx context.Context, incomingDocID uuid.UUID) (*domain.RouteStep, error)
	FindNextStep(ctx context.Context, outgoingDocID uuid.UUID, currentOrder int) (*domain.RouteStep, error)
	AttachIncomingDoc(ctx context.Context, routeID, incomingDocID uuid.UUID) error
	// UpdateStatus propagates a rejection upstream to the parent outgoing doc.
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, approverID *uuid.UUID) error
}

// Service defines the interface for incoming document business logic
type Service interface {
	GetAllIncomingDocs(ctx context.Context, page, limit int) ([]domain.IncomingDocResponse, int, error)
	GetAllIncomingDocsExcludingSender(ctx context.Context, senderID uuid.UUID, page, limit int) ([]domain.IncomingDocResponse, int, error)
	GetIncomingDocByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDocResponse, error)
	GetIncomingDocsByReceiver(ctx context.Context, receiverID uuid.UUID) ([]domain.IncomingDocResponse, error)
	GetIncomingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.IncomingDocResponse, int, error)
	GetIncomingDocsByStatus(ctx context.Context, status string) ([]domain.IncomingDocResponse, error)
	ReceiveDocument(ctx context.Context, req domain.ReceiveDocumentRequest) (*domain.IncomingDocResponse, error)
	ApproveDocument(ctx context.Context, id uuid.UUID, req domain.ApproveDocumentRequest) (*domain.IncomingDocResponse, error)
	// CreateIncomingDocForStep creates the incoming doc for a single route step
	// (status pending) and links it back to the route.
	CreateIncomingDocForStep(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, step *domain.RouteStep, remark string) error
	// CreateDirectIncomingDocs creates incoming docs directly for the given
	// departments without a sequential route (legacy standalone "incoming" upload).
	CreateDirectIncomingDocs(ctx context.Context, docDetailsID uuid.UUID, outgoingDocID uuid.UUID, creatorID *uuid.UUID, deptIDs []uuid.UUID, remark string) error
}

type service struct {
	repo   Repository
	routes RouteAdvancer
}

// NewService creates a new incoming document service
func NewService(repo Repository, routes RouteAdvancer) Service {
	return &service{
		repo:   repo,
		routes: routes,
	}
}

func (s *service) GetAllIncomingDocs(ctx context.Context, page, limit int) ([]domain.IncomingDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get all incoming documents", err)
	}

	responses := make([]domain.IncomingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, total, nil
}

func (s *service) GetAllIncomingDocsExcludingSender(ctx context.Context, senderID uuid.UUID, page, limit int) ([]domain.IncomingDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindAllExcludingSender(ctx, senderID, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get all incoming documents excluding sender", err)
	}

	responses := make([]domain.IncomingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, total, nil
}

func (s *service) GetIncomingDocByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("IncomingDoc", id.String())
	}

	response := doc.ToResponse()
	return &response, nil
}

func (s *service) GetIncomingDocsByReceiver(ctx context.Context, receiverID uuid.UUID) ([]domain.IncomingDocResponse, error) {
	docs, err := s.repo.FindByReceiverID(ctx, receiverID)
	if err != nil {
		return nil, util.NewDatabaseError("get incoming documents by receiver", err)
	}

	responses := make([]domain.IncomingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, nil
}

func (s *service) GetIncomingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.IncomingDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindByDepartmentID(ctx, deptID, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get incoming documents by department", err)
	}

	responses := make([]domain.IncomingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, total, nil
}

func (s *service) GetIncomingDocsByStatus(ctx context.Context, status string) ([]domain.IncomingDocResponse, error) {
	docs, err := s.repo.FindByStatus(ctx, status)
	if err != nil {
		return nil, util.NewDatabaseError("get incoming documents by status", err)
	}

	responses := make([]domain.IncomingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, nil
}

func (s *service) ReceiveDocument(ctx context.Context, req domain.ReceiveDocumentRequest) (*domain.IncomingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, req.IncomingDocID)
	if err != nil {
		return nil, util.NewNotFoundError("IncomingDoc", req.IncomingDocID.String())
	}

	if doc.Status != domain.IncomingStatusPending {
		return nil, util.NewValidationError(fmt.Sprintf("cannot receive a document with status '%s'", doc.Status))
	}

	now := time.Now()
	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		return nil, util.NewInvalidInputError("receiver_id", "must be a valid UUID")
	}

	doc.Status = domain.IncomingStatusReceived
	doc.ReceivedDate = &now
	doc.UpdatedBy = &receiverID
	if req.Remark != "" {
		doc.Remark = req.Remark
	}

	if err := s.repo.Update(ctx, doc.ID, doc); err != nil {
		return nil, util.NewDatabaseError("receive document", err)
	}

	updated, err := s.repo.FindByID(ctx, doc.ID)
	if err != nil {
		return nil, util.NewDatabaseError("fetch updated document", err)
	}

	resp := updated.ToResponse()
	return &resp, nil
}

// ApproveDocument approves or rejects an incoming document. On approval it
// advances the outgoing document's flow by lazily creating the next route
// step's incoming document. On rejection the flow stops.
func (s *service) ApproveDocument(ctx context.Context, id uuid.UUID, req domain.ApproveDocumentRequest) (*domain.IncomingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("IncomingDoc", id.String())
	}

	if doc.Status != domain.IncomingStatusReceived {
		return nil, util.NewValidationError(fmt.Sprintf("cannot approve/reject a document with status '%s' (must be 'received')", doc.Status))
	}

	approved := req.Status == string(domain.IncomingStatusApproved)
	rejected := req.Status == string(domain.IncomingStatusRejected)
	if !approved && !rejected {
		return nil, util.NewInvalidInputError("status", "must be 'approved' or 'rejected'")
	}

	now := time.Now()
	if req.ApproverID != "" {
		approverID, perr := uuid.Parse(req.ApproverID)
		if perr != nil {
			return nil, util.NewInvalidInputError("approver_id", "must be a valid UUID")
		}
		doc.ApproverID = &approverID
		doc.UpdatedBy = &approverID
	}
	doc.ApproverDate = &now
	if approved {
		doc.Status = domain.IncomingStatusApproved
	} else {
		doc.Status = domain.IncomingStatusRejected
	}
	if req.Remark != "" {
		doc.Remark = req.Remark
	}

	if err := s.repo.Update(ctx, doc.ID, doc); err != nil {
		return nil, util.NewDatabaseError("approve document", err)
	}

	if approved {
		// Advance the flow to the next department (if any).
		if err := s.advanceFlow(ctx, doc); err != nil {
			log.Error().Err(err).Str("incoming_doc_id", doc.ID.String()).Msg("failed to advance flow to next step")
		}
	}

	if rejected && s.routes != nil {
		// Propagate the rejection to the parent outgoing doc so its status
		// reflects "rejected" in the outgoing-docs table.
		if step, stepErr := s.routes.FindRouteByIncomingDocID(ctx, doc.ID); stepErr == nil && step != nil {
			if updateErr := s.routes.UpdateStatus(ctx, step.OutgoingDocID, domain.OutgoingStatusRejected, doc.ApproverID); updateErr != nil {
				log.Error().Err(updateErr).Str("outgoing_doc_id", step.OutgoingDocID.String()).Msg("failed to propagate rejection to outgoing doc")
			}
		}
	}

	updated, err := s.repo.FindByID(ctx, doc.ID)
	if err != nil {
		return nil, util.NewDatabaseError("fetch updated document", err)
	}
	resp := updated.ToResponse()
	return &resp, nil
}

// advanceFlow creates the next route step's incoming document after approval.
func (s *service) advanceFlow(ctx context.Context, approvedDoc *domain.IncomingDoc) error {
	if s.routes == nil {
		return fmt.Errorf("route advancer not configured")
	}
	step, err := s.routes.FindRouteByIncomingDocID(ctx, approvedDoc.ID)
	if err != nil {
		return err
	}
	if step == nil {
		return nil // not part of a routed flow
	}
	next, err := s.routes.FindNextStep(ctx, step.OutgoingDocID, step.SequenceOrder)
	if err != nil {
		return err
	}
	if next == nil {
		return nil // flow complete — this was the last step
	}
	if next.IncomingDocID != nil {
		return nil // already created (idempotent)
	}
	nextStep := *next
	return s.CreateIncomingDocForStep(ctx, approvedDoc.DocDetailsID, approvedDoc.CreatedBy, &nextStep, approvedDoc.Remark)
}

// CreateIncomingDocForStep creates the incoming doc for a route step and links it back.
func (s *service) CreateIncomingDocForStep(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, step *domain.RouteStep, remark string) error {
	now := time.Now()
	outgoingDocID := step.OutgoingDocID
	deptID := step.DeptID
	doc := &domain.IncomingDoc{
		DocDetailsID:  docDetailsID,
		OutgoingDocID: &outgoingDocID,
		CreatedBy:     creatorID,
		Status:        domain.IncomingStatusPending,
		Remark:        remark,
		IncomingDate:  &now,
		DeptID:        &deptID,
	}
	if err := s.repo.Create(ctx, doc); err != nil {
		return fmt.Errorf("failed to create incoming doc for dept %s: %w", deptID, err)
	}
	if s.routes != nil {
		if err := s.routes.AttachIncomingDoc(ctx, step.ID, doc.ID); err != nil {
			return fmt.Errorf("failed to attach incoming doc to route step: %w", err)
		}
	}
	return nil
}

// CreateDirectIncomingDocs creates incoming docs directly (no route) for each dept.
func (s *service) CreateDirectIncomingDocs(ctx context.Context, docDetailsID uuid.UUID, outgoingDocID uuid.UUID, creatorID *uuid.UUID, deptIDs []uuid.UUID, remark string) error {
	now := time.Now()
	var outgoingDocIDPtr *uuid.UUID
	if outgoingDocID != uuid.Nil {
		outgoingDocIDPtr = &outgoingDocID
	}
	for _, deptID := range deptIDs {
		deptIDCopy := deptID
		doc := &domain.IncomingDoc{
			DocDetailsID:  docDetailsID,
			OutgoingDocID: outgoingDocIDPtr,
			CreatedBy:     creatorID,
			Status:        domain.IncomingStatusPending,
			Remark:        remark,
			IncomingDate:  &now,
			DeptID:        &deptIDCopy,
		}
		if err := s.repo.Create(ctx, doc); err != nil {
			return fmt.Errorf("failed to create incoming doc for dept %s: %w", deptID, err)
		}
	}
	return nil
}
