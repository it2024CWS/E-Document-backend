package outgoingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"github.com/google/uuid"
)

// Repository defines the interface for outgoing document data access
type Repository interface {
	FindAll(ctx context.Context, limit, offset int) ([]domain.OutgoingDoc, int, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDoc, error)
	FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int) ([]domain.OutgoingDoc, int, error)
	FindRecipientsByOutgoingDocID(ctx context.Context, outgoingDocID uuid.UUID) ([]domain.RecipientInfo, error)
	Create(ctx context.Context, doc *domain.OutgoingDoc) error
	// UpdateStatus sets the owner-approval gate status and the approver (updated_by).
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, approverID *uuid.UUID) error

	// Route operations — drive the sequential department flow.
	CreateRoutes(ctx context.Context, routes []domain.OutgoingDocRoute) error
	FindRouteByIncomingDocID(ctx context.Context, incomingDocID uuid.UUID) (*domain.RouteStep, error)
	FindNextStep(ctx context.Context, outgoingDocID uuid.UUID, currentOrder int) (*domain.RouteStep, error)
	AttachIncomingDoc(ctx context.Context, routeID, incomingDocID uuid.UUID) error
}
