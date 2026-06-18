package incomingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
)

// --- stub repository (shared with approve_test.go) ---

type stubRepo struct {
	created  []*domain.IncomingDoc
	updated  []*domain.IncomingDoc
	findByID func(id uuid.UUID) (*domain.IncomingDoc, error)
}

func (s *stubRepo) Create(ctx context.Context, doc *domain.IncomingDoc) error {
	doc.ID = uuid.New()
	s.created = append(s.created, doc)
	return nil
}
func (s *stubRepo) Update(ctx context.Context, id uuid.UUID, doc *domain.IncomingDoc) error {
	s.updated = append(s.updated, doc)
	return nil
}
func (s *stubRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDoc, error) {
	return s.findByID(id)
}
func (s *stubRepo) FindAll(ctx context.Context, limit, offset int) ([]domain.IncomingDoc, int, error) {
	return nil, 0, nil
}
func (s *stubRepo) FindByReceiverID(ctx context.Context, id uuid.UUID) ([]domain.IncomingDoc, error) {
	return nil, nil
}
func (s *stubRepo) FindByDepartmentID(ctx context.Context, id uuid.UUID, limit, offset int) ([]domain.IncomingDoc, int, error) {
	return nil, 0, nil
}
func (s *stubRepo) FindByStatus(ctx context.Context, status string) ([]domain.IncomingDoc, error) {
	return nil, nil
}
func (s *stubRepo) FindByDocID(ctx context.Context, id uuid.UUID) ([]domain.IncomingDoc, error) {
	return nil, nil
}
func (s *stubRepo) FindByOutgoingDocID(ctx context.Context, id uuid.UUID) ([]domain.IncomingDoc, error) {
	return nil, nil
}
func (s *stubRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error { return nil }
func (s *stubRepo) FindAllExcludingSender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.IncomingDoc, int, error) {
	return nil, 0, nil
}

// --- tests ---

// TestCreateIncomingDocs_StatusPending ensures direct incoming docs start as pending.
func TestCreateIncomingDocs_StatusPending(t *testing.T) {
	repo := &stubRepo{}
	svc := &service{repo: repo}

	docID := uuid.New()
	creatorID := uuid.New()
	deptID1, deptID2 := uuid.New(), uuid.New()

	outgoingDocID := uuid.New()
	err := svc.CreateDirectIncomingDocs(context.Background(), docID, outgoingDocID, &creatorID,
		[]uuid.UUID{deptID1, deptID2}, "test remark")
	if err != nil {
		t.Fatalf("CreateDirectIncomingDocs error: %v", err)
	}

	if len(repo.created) != 2 {
		t.Fatalf("expected 2 incoming docs created, got %d", len(repo.created))
	}

	for i, doc := range repo.created {
		if doc.Status != domain.IncomingStatusPending {
			t.Errorf("incoming_doc[%d]: expected status=pending, got %s", i, doc.Status)
		}
	}
}

// TestReceiveDocument_SetsReceived ensures receiving a pending doc sets status + received_date.
func TestReceiveDocument_SetsReceived(t *testing.T) {
	now := time.Now()
	docID := uuid.New()
	receiverID := uuid.New()

	pendingDoc := &domain.IncomingDoc{
		ID:           docID,
		Status:       domain.IncomingStatusPending,
		DocDetailsID: uuid.New(),
		IncomingDate: &now,
	}

	repo := &stubRepo{
		findByID: func(id uuid.UUID) (*domain.IncomingDoc, error) {
			return pendingDoc, nil
		},
	}
	svc := &service{repo: repo}

	req := domain.ReceiveDocumentRequest{
		IncomingDocID: docID,
		ReceiverID:    receiverID.String(),
		Remark:        "physically received",
	}

	if _, err := svc.ReceiveDocument(context.Background(), req); err != nil {
		t.Fatalf("ReceiveDocument error: %v", err)
	}

	if len(repo.updated) == 0 {
		t.Fatal("expected Update to be called")
	}
	updatedDoc := repo.updated[0]

	if updatedDoc.Status != domain.IncomingStatusReceived {
		t.Errorf("expected status=received, got %s", updatedDoc.Status)
	}
	if updatedDoc.ReceivedDate == nil {
		t.Error("received_date should be set")
	}
}

// TestReceiveDocument_RejectNonPending ensures non-pending docs cannot be received again.
func TestReceiveDocument_RejectNonPending(t *testing.T) {
	docID := uuid.New()
	now := time.Now()

	receivedDoc := &domain.IncomingDoc{
		ID:           docID,
		Status:       domain.IncomingStatusReceived,
		ReceivedDate: &now,
	}

	repo := &stubRepo{
		findByID: func(id uuid.UUID) (*domain.IncomingDoc, error) {
			return receivedDoc, nil
		},
	}
	svc := &service{repo: repo}

	_, err := svc.ReceiveDocument(context.Background(), domain.ReceiveDocumentRequest{
		IncomingDocID: docID,
		ReceiverID:    uuid.New().String(),
	})

	if err == nil {
		t.Error("FAIL: expected error when receiving an already-received document")
	}

	if len(repo.updated) > 0 {
		t.Error("Update should NOT have been called for a non-pending doc")
	}
}
