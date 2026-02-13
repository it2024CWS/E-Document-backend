package outgoingdoc

import (
	"context"
	"e-document-backend/internal/domain"
)

// Repository defines the interface for outgoing document data access
type Repository interface {
	FindAll(ctx context.Context) ([]domain.OutgoingDoc, error)
	FindByID(ctx context.Context, id int) (*domain.OutgoingDoc, error)
	FindByUserID(ctx context.Context, userID string) ([]domain.OutgoingDoc, error)
	Create(ctx context.Context, doc *domain.OutgoingDoc) error
}
