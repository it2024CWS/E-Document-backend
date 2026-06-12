package incomingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"testing"
	"time"

	"github.com/google/uuid"
)

// --- stub repository ---

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

// --- tests ---

// TestCreateIncomingDocs_NoIncomingNoGenerated ensures incoming_no is NOT set during creation
func TestCreateIncomingDocs_NoIncomingNoGenerated(t *testing.T) {
	repo := &stubRepo{}
	svc := &service{repo: repo}

	docID := uuid.New()
	creatorID := uuid.New()
	deptID1, deptID2 := uuid.New(), uuid.New()

	outgoingDocID := uuid.New()
	err := svc.CreateIncomingDocs(context.Background(), docID, outgoingDocID, &creatorID,
		[]uuid.UUID{deptID1, deptID2}, "test remark")
	if err != nil {
		t.Fatalf("CreateIncomingDocs error: %v", err)
	}

	if len(repo.created) != 2 {
		t.Fatalf("expected 2 incoming docs created, got %d", len(repo.created))
	}

	for i, doc := range repo.created {
		if doc.IncomingNo != "" {
			t.Errorf("incoming_doc[%d]: incoming_no should be empty at creation, got %q", i, doc.IncomingNo)
		} else {
			t.Logf("PASS incoming_doc[%d]: incoming_no is empty ✓", i)
		}

		if doc.Status != domain.IncomingStatusPending {
			t.Errorf("incoming_doc[%d]: expected status=pending, got %s", i, doc.Status)
		}
	}
}

// TestReceiveDocument_IncomingNoGeneratedOnReceive ensures incoming_no IS generated on receive
func TestReceiveDocument_IncomingNoGeneratedOnReceive(t *testing.T) {
	now := time.Now()
	docID := uuid.New()
	receiverID := uuid.New()

	pendingDoc := &domain.IncomingDoc{
		ID:           docID,
		IncomingNo:   "", // null in DB — not yet received
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

	resp, err := svc.ReceiveDocument(context.Background(), req)
	if err != nil {
		t.Fatalf("ReceiveDocument error: %v", err)
	}

	// Verify the update call carried a generated incoming_no
	if len(repo.updated) == 0 {
		t.Fatal("expected Update to be called")
	}

	updatedDoc := repo.updated[0]

	if updatedDoc.IncomingNo == "" {
		t.Error("FAIL: incoming_no is still empty after ReceiveDocument — should have been generated")
	} else {
		t.Logf("PASS: incoming_no=%q generated on receive ✓", updatedDoc.IncomingNo)
	}

	if updatedDoc.Status != domain.IncomingStatusReceived {
		t.Errorf("expected status=received, got %s", updatedDoc.Status)
	}

	if updatedDoc.ReceivedDate == nil {
		t.Error("received_date should be set")
	}

	// Verify format matches our number generator
	if !isValidIncomingNo(updatedDoc.IncomingNo) {
		t.Errorf("incoming_no format unexpected: %q", updatedDoc.IncomingNo)
	}

	_ = resp
}

// TestReceiveDocument_RejectNonPending ensures non-pending docs cannot be received again
func TestReceiveDocument_RejectNonPending(t *testing.T) {
	docID := uuid.New()
	now := time.Now()
	existingNo := util.GenerateIncomingNumber()

	receivedDoc := &domain.IncomingDoc{
		ID:         docID,
		IncomingNo: existingNo,
		Status:     domain.IncomingStatusReceived,
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
	} else {
		t.Logf("PASS: correctly rejected with error: %v ✓", err)
	}

	if len(repo.updated) > 0 {
		t.Error("Update should NOT have been called for a non-pending doc")
	}
}

func isValidIncomingNo(s string) bool {
	// GenerateIncomingNumber produces e.g. "IN13062026010203123"
	return len(s) > 5 && s[:2] == "IN"
}
