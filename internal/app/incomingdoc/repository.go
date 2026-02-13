package incomingdoc

import (
	"context"
	"e-document-backend/internal/domain"
)

// Repository defines the interface for incoming document data access
type Repository interface {
	FindAll(ctx context.Context) ([]domain.IncomingDoc, error)
	FindByID(ctx context.Context, id int) (*domain.IncomingDoc, error)
	FindByReceiverID(ctx context.Context, receiverID string) ([]domain.IncomingDoc, error)
	FindByStatus(ctx context.Context, status string) ([]domain.IncomingDoc, error)
	Create(ctx context.Context, doc *domain.IncomingDoc) error
	Update(ctx context.Context, id int, doc *domain.IncomingDoc) error
	UpdateStatus(ctx context.Context, id int, status string) error
}
