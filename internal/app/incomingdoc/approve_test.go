package incomingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
)

// --- stub route advancer ---

type stubRoutes struct {
	routeByIncoming *domain.RouteStep
	nextStep        *domain.RouteStep
	attached        []uuid.UUID // route ids that got an incoming doc attached
}

func (s *stubRoutes) FindRouteByIncomingDocID(ctx context.Context, incomingDocID uuid.UUID) (*domain.RouteStep, error) {
	return s.routeByIncoming, nil
}
func (s *stubRoutes) FindNextStep(ctx context.Context, outgoingDocID uuid.UUID, currentOrder int) (*domain.RouteStep, error) {
	return s.nextStep, nil
}
func (s *stubRoutes) AttachIncomingDoc(ctx context.Context, routeID, incomingDocID uuid.UUID) error {
	s.attached = append(s.attached, routeID)
	return nil
}

// receivedDoc builds an incoming doc in "received" state ready to be approved.
func receivedDoc() *domain.IncomingDoc {
	now := time.Now()
	return &domain.IncomingDoc{
		ID:           uuid.New(),
		Status:       domain.IncomingStatusReceived,
		DocDetailsID: uuid.New(),
		IncomingDate: &now,
		ReceivedDate: &now,
	}
}

// TestApprove_NonLastStep_CreatesNext: approving a received doc that has a next
// route step must create the next department's incoming doc and link it.
func TestApprove_NonLastStep_CreatesNext(t *testing.T) {
	doc := receivedDoc()
	repo := &stubRepo{findByID: func(id uuid.UUID) (*domain.IncomingDoc, error) { return doc, nil }}

	outID := uuid.New()
	nextRouteID := uuid.New()
	routes := &stubRoutes{
		routeByIncoming: &domain.RouteStep{ID: uuid.New(), OutgoingDocID: outID, SequenceOrder: 1},
		nextStep:        &domain.RouteStep{ID: nextRouteID, OutgoingDocID: outID, DeptID: uuid.New(), SequenceOrder: 2},
	}
	svc := &service{repo: repo, routes: routes}

	_, err := svc.ApproveDocument(context.Background(), doc.ID, domain.ApproveDocumentRequest{
		ApproverID: uuid.New().String(),
		Status:     string(domain.IncomingStatusApproved),
		Remark:     "ok",
	})
	if err != nil {
		t.Fatalf("ApproveDocument error: %v", err)
	}
	if doc.Status != domain.IncomingStatusApproved {
		t.Errorf("expected status approved, got %s", doc.Status)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 next incoming doc created, got %d", len(repo.created))
	}
	if len(routes.attached) != 1 || routes.attached[0] != nextRouteID {
		t.Errorf("expected next route %s to get incoming doc attached, got %v", nextRouteID, routes.attached)
	}
}

// TestApprove_LastStep_Completes: approving the final step must NOT create another doc.
func TestApprove_LastStep_Completes(t *testing.T) {
	doc := receivedDoc()
	repo := &stubRepo{findByID: func(id uuid.UUID) (*domain.IncomingDoc, error) { return doc, nil }}
	routes := &stubRoutes{
		routeByIncoming: &domain.RouteStep{ID: uuid.New(), OutgoingDocID: uuid.New(), SequenceOrder: 3},
		nextStep:        nil, // no next step
	}
	svc := &service{repo: repo, routes: routes}

	_, err := svc.ApproveDocument(context.Background(), doc.ID, domain.ApproveDocumentRequest{
		Status: string(domain.IncomingStatusApproved),
	})
	if err != nil {
		t.Fatalf("ApproveDocument error: %v", err)
	}
	if len(repo.created) != 0 {
		t.Errorf("expected no incoming doc created on last step, got %d", len(repo.created))
	}
	if len(routes.attached) != 0 {
		t.Errorf("expected no attach on last step, got %d", len(routes.attached))
	}
}

// TestApprove_Rejected_StopsFlow: rejecting must set status rejected and create nothing.
func TestApprove_Rejected_StopsFlow(t *testing.T) {
	doc := receivedDoc()
	repo := &stubRepo{findByID: func(id uuid.UUID) (*domain.IncomingDoc, error) { return doc, nil }}
	routes := &stubRoutes{
		routeByIncoming: &domain.RouteStep{ID: uuid.New(), OutgoingDocID: uuid.New(), SequenceOrder: 1},
		nextStep:        &domain.RouteStep{ID: uuid.New(), OutgoingDocID: uuid.New(), SequenceOrder: 2},
	}
	svc := &service{repo: repo, routes: routes}

	_, err := svc.ApproveDocument(context.Background(), doc.ID, domain.ApproveDocumentRequest{
		Status: string(domain.IncomingStatusRejected),
		Remark: "missing signature",
	})
	if err != nil {
		t.Fatalf("ApproveDocument error: %v", err)
	}
	if doc.Status != domain.IncomingStatusRejected {
		t.Errorf("expected status rejected, got %s", doc.Status)
	}
	if len(repo.created) != 0 {
		t.Errorf("expected no next doc created on rejection, got %d", len(repo.created))
	}
}

// TestApprove_NotReceived_Errors: a pending (not received) doc cannot be approved.
func TestApprove_NotReceived_Errors(t *testing.T) {
	doc := receivedDoc()
	doc.Status = domain.IncomingStatusPending
	repo := &stubRepo{findByID: func(id uuid.UUID) (*domain.IncomingDoc, error) { return doc, nil }}
	svc := &service{repo: repo, routes: &stubRoutes{}}

	_, err := svc.ApproveDocument(context.Background(), doc.ID, domain.ApproveDocumentRequest{
		Status: string(domain.IncomingStatusApproved),
	})
	if err == nil {
		t.Fatal("expected error approving a non-received document")
	}
	if len(repo.updated) != 0 {
		t.Errorf("expected no update for invalid transition, got %d", len(repo.updated))
	}
}

// TestApprove_InvalidStatus_Errors: status other than approved/rejected is rejected.
func TestApprove_InvalidStatus_Errors(t *testing.T) {
	doc := receivedDoc()
	repo := &stubRepo{findByID: func(id uuid.UUID) (*domain.IncomingDoc, error) { return doc, nil }}
	svc := &service{repo: repo, routes: &stubRoutes{}}

	_, err := svc.ApproveDocument(context.Background(), doc.ID, domain.ApproveDocumentRequest{
		Status: "maybe",
	})
	if err == nil {
		t.Fatal("expected error for invalid status value")
	}
}
