package outgoingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// StepCreator creates the incoming doc for a single route step. Implemented by
// the incomingdoc service; injected so the owner-approval gate can lazily kick
// off the recipient flow without a circular package dependency.
type StepCreator interface {
	CreateIncomingDocForStep(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, step *domain.RouteStep, remark string) error
}

type Service interface {
	GetAllOutgoingDocs(ctx context.Context, page, limit int) ([]domain.OutgoingDocResponse, int, error)
	GetOutgoingDocByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDocResponse, error)
	GetOutgoingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.OutgoingDocResponse, int, error)
	CreateOutgoingDocWithParams(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, deptID *uuid.UUID) (uuid.UUID, error)
	// CreateOutgoingDocWithRoute creates the outgoing doc (status pending, awaiting
	// owner approval) and its ordered recipient route steps. No incoming docs are
	// created — the recipient flow starts only when the owner head approves.
	CreateOutgoingDocWithRoute(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, orderedDeptIDs []uuid.UUID) (uuid.UUID, error)
	// ApproveOutgoingDoc is the owner department head's gate: on approval the first
	// recipient step's incoming doc is created; on rejection the flow stops.
	ApproveOutgoingDoc(ctx context.Context, id uuid.UUID, req domain.ApproveDocumentRequest) (*domain.OutgoingDocResponse, error)
}

type service struct {
	repo        Repository
	stepCreator StepCreator
}

func NewService(repo Repository, stepCreator StepCreator) Service {
	return &service{repo: repo, stepCreator: stepCreator}
}

// buildStatusCounts computes status counts from a recipients slice.
// "waiting" steps are counted in Total but not in any status bucket.
func buildStatusCounts(recipients []domain.RecipientInfo) domain.StatusCounts {
	counts := domain.StatusCounts{Total: len(recipients)}
	for _, r := range recipients {
		switch r.Status {
		case string(domain.IncomingStatusPending):
			counts.Pending++
		case string(domain.IncomingStatusReceived):
			counts.Received++
		case string(domain.IncomingStatusApproved):
			counts.Approved++
		case string(domain.IncomingStatusRejected):
			counts.Rejected++
		}
	}
	return counts
}

// deriveFlow computes the current location and overall flow status from the
// ordered recipients. Rejected anywhere → rejected; all approved → completed;
// otherwise in_progress sitting at the first non-approved step.
func deriveFlow(resp *domain.OutgoingDocResponse, recipients []domain.RecipientInfo) {
	resp.FlowStatus = domain.FlowStatusCompleted
	for _, r := range recipients {
		if r.Status == string(domain.IncomingStatusRejected) {
			resp.FlowStatus = domain.FlowStatusRejected
			resp.CurrentDepartment = r.DepartmentName
			resp.CurrentStatus = r.Status
			return
		}
	}
	for _, r := range recipients {
		if r.IsCurrent {
			resp.FlowStatus = domain.FlowStatusInProgress
			resp.CurrentDepartment = r.DepartmentName
			if r.Status == domain.StatusWaiting {
				resp.CurrentStatus = string(domain.IncomingStatusPending)
			} else {
				resp.CurrentStatus = r.Status
			}
			return
		}
	}
	// No current step → all approved (completed). Point at the last step.
	if n := len(recipients); n > 0 {
		resp.CurrentDepartment = recipients[n-1].DepartmentName
	}
	resp.CurrentStatus = ""
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
	deriveFlow(resp, recipients)

	// The owner-approval gate overrides the recipient-derived flow until the
	// owner department head has acted. While pending there are no recipient
	// incoming docs yet, so deriveFlow would wrongly report "completed".
	switch resp.Status {
	case domain.OutgoingStatusPending:
		resp.FlowStatus = domain.FlowStatusPendingApproval
		resp.CurrentStatus = domain.OutgoingStatusPending
		resp.CurrentDepartment = ""
	case domain.OutgoingStatusRejected:
		resp.FlowStatus = domain.FlowStatusRejected
		resp.CurrentStatus = domain.OutgoingStatusRejected
		resp.CurrentDepartment = ""
	}
}

// BuildRoute returns the ordered recipient-department list for a flow:
// the selected departments (in order, de-duplicated), then Director Office last
// if provided. The owner department is NOT part of the route — the owner
// department head approves the document on the outgoing side (outgoing_docs.status),
// which gates the start of this recipient flow. ownerDept is still used to exclude
// the owner if it was accidentally included in the selection.
func BuildRoute(ownerDept uuid.UUID, selected []uuid.UUID, director *uuid.UUID) []uuid.UUID {
	route := make([]uuid.UUID, 0, len(selected)+1)
	seen := map[uuid.UUID]bool{}

	if ownerDept != uuid.Nil {
		seen[ownerDept] = true // exclude owner from the recipient route
	}
	for _, d := range selected {
		if d == uuid.Nil || seen[d] {
			continue
		}
		route = append(route, d)
		seen[d] = true
	}
	if director != nil && *director != uuid.Nil && !seen[*director] {
		route = append(route, *director)
		seen[*director] = true
	}
	return route
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
		DocDetailsID: docDetailsID,
		CreatedBy:    creatorID,
	}
	if err := s.repo.Create(ctx, doc); err != nil {
		return uuid.Nil, util.NewDatabaseError("create outgoing document", err)
	}
	return doc.ID, nil
}

// CreateOutgoingDocWithRoute creates the outgoing doc (status pending) and inserts
// its ordered recipient route steps. It does NOT create any incoming docs — the
// recipient flow starts only when the owner department head approves the doc.
func (s *service) CreateOutgoingDocWithRoute(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, orderedDeptIDs []uuid.UUID) (uuid.UUID, error) {
	doc := &domain.OutgoingDoc{
		DocDetailsID: docDetailsID,
		CreatedBy:    creatorID,
		Status:       domain.OutgoingStatusPending,
	}
	if err := s.repo.Create(ctx, doc); err != nil {
		return uuid.Nil, util.NewDatabaseError("create outgoing document", err)
	}

	routes := make([]domain.OutgoingDocRoute, 0, len(orderedDeptIDs))
	for i, deptID := range orderedDeptIDs {
		routes = append(routes, domain.OutgoingDocRoute{
			OutgoingDocID: doc.ID,
			DeptID:        deptID,
			SequenceOrder: i + 1,
		})
	}
	if err := s.repo.CreateRoutes(ctx, routes); err != nil {
		return uuid.Nil, util.NewDatabaseError("create outgoing routes", err)
	}

	return doc.ID, nil
}

// ApproveOutgoingDoc is the owner department head's approval gate. It must be
// called while the doc is still pending. On approval it creates the first
// recipient step's incoming doc (kicking off the sequential flow); on rejection
// it stops the flow and no recipient docs are created.
func (s *service) ApproveOutgoingDoc(ctx context.Context, id uuid.UUID, req domain.ApproveDocumentRequest) (*domain.OutgoingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("OutgoingDoc", id.String())
	}

	if doc.Status != domain.OutgoingStatusPending {
		return nil, util.NewValidationError(fmt.Sprintf("cannot approve/reject an outgoing document with status '%s' (must be 'pending')", doc.Status))
	}

	approved := req.Status == domain.OutgoingStatusApproved
	rejected := req.Status == domain.OutgoingStatusRejected
	if !approved && !rejected {
		return nil, util.NewInvalidInputError("status", "must be 'approved' or 'rejected'")
	}

	var approverID *uuid.UUID
	if req.ApproverID != "" {
		parsed, perr := uuid.Parse(req.ApproverID)
		if perr != nil {
			return nil, util.NewInvalidInputError("approver_id", "must be a valid UUID")
		}
		approverID = &parsed
	}

	newStatus := domain.OutgoingStatusApproved
	if rejected {
		newStatus = domain.OutgoingStatusRejected
	}
	if err := s.repo.UpdateStatus(ctx, id, newStatus, approverID); err != nil {
		return nil, util.NewDatabaseError("update outgoing document status", err)
	}

	// On approval, start the recipient flow by creating the first step's incoming doc.
	if approved {
		if err := s.startRecipientFlow(ctx, doc, req.Remark); err != nil {
			log.Error().Err(err).Str("outgoing_doc_id", id.String()).Msg("failed to start recipient flow after approval")
		}
	}

	return s.GetOutgoingDocByID(ctx, id)
}

// startRecipientFlow creates the first route step's incoming doc after the owner
// department head approves the outgoing document.
func (s *service) startRecipientFlow(ctx context.Context, doc *domain.OutgoingDoc, remark string) error {
	if s.stepCreator == nil {
		return fmt.Errorf("step creator not configured")
	}
	// The first recipient is sequence_order 1 (FindNextStep after order 0).
	first, err := s.repo.FindNextStep(ctx, doc.ID, 0)
	if err != nil {
		return err
	}
	if first == nil {
		return nil // no recipients configured — nothing to start
	}
	if first.IncomingDocID != nil {
		return nil // already created (idempotent)
	}
	step := *first
	return s.stepCreator.CreateIncomingDocForStep(ctx, doc.DocDetailsID, doc.CreatedBy, &step, remark)
}
