package incomingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"github.com/google/uuid"
)

// Repository defines the interface for incoming document data access
type Repository interface {
	FindAll(ctx context.Context, limit, offset int) ([]domain.IncomingDoc, int, error)
	FindAllExcludingSender(ctx context.Context, senderID uuid.UUID, limit, offset int) ([]domain.IncomingDoc, int, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDoc, error)
	FindByReceiverID(ctx context.Context, receiverID uuid.UUID) ([]domain.IncomingDoc, error)
	FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int) ([]domain.IncomingDoc, int, error)
	FindByStatus(ctx context.Context, status string) ([]domain.IncomingDoc, error)
	FindByDocID(ctx context.Context, docID uuid.UUID) ([]domain.IncomingDoc, error)
	FindByOutgoingDocID(ctx context.Context, outgoingDocID uuid.UUID) ([]domain.IncomingDoc, error)
	Create(ctx context.Context, doc *domain.IncomingDoc) error
	Update(ctx context.Context, id uuid.UUID, doc *domain.IncomingDoc) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}
