package outgoingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"time"

	"github.com/google/uuid"
)

// DocFilter holds optional query filters for outgoing document listing.
type DocFilter struct {
	DocNo     string
	Status    string
	StartDate *time.Time
	EndDate   *time.Time // exclusive upper bound (caller adds 24h)
}

// Repository defines the interface for outgoing document data access
type Repository interface {
	FindAll(ctx context.Context, limit, offset int, filter DocFilter) ([]domain.OutgoingDoc, int, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDoc, error)
	FindByDepartmentID(ctx context.Context, deptID uuid.UUID, limit, offset int, filter DocFilter) ([]domain.OutgoingDoc, int, error)
	FindRecipientsByOutgoingDocID(ctx context.Context, outgoingDocID uuid.UUID) ([]domain.RecipientInfo, error)
	Create(ctx context.Context, doc *domain.OutgoingDoc) error
	// UpdateStatus sets the owner-approval gate status and the approver (updated_by).
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, approverID *uuid.UUID) error
	// UpdateDocMeta patches doc_no / doc_name on the linked doc_details row, and
	// bumps outgoing_docs.updated_by. Empty strings are skipped.
	UpdateDocMeta(ctx context.Context, id uuid.UUID, docNo, docName string, updaterID *uuid.UUID) error
	// SoftDelete marks both outgoing_docs and its linked doc_details as deleted.
	SoftDelete(ctx context.Context, id uuid.UUID, updaterID *uuid.UUID) error
	// FindByDocDetailsID returns the outgoing_doc row whose doc_details_id matches.
	// Returns nil (no error) when not found.
	FindByDocDetailsID(ctx context.Context, docDetailsID uuid.UUID) (*domain.OutgoingDoc, error)
	// ReplaceFile updates the versions row for the outgoing doc's doc_details_id
	// so it points at a new stored file. Used by the edit flow when the owner
	// uploads a replacement file.
	ReplaceFile(ctx context.Context, outgoingDocID uuid.UUID, docPath, fileType string, updaterID *uuid.UUID) error

	// Route operations — drive the sequential department flow.
	CreateRoutes(ctx context.Context, routes []domain.OutgoingDocRoute) error
	// ReplaceRoutes clears the outgoing_doc_routes for a document and rewrites them
	// from deptIDs in order. Safe only while no route step has an incoming_doc_id
	// (i.e. before the owner-approval gate opens the recipient flow).
	ReplaceRoutes(ctx context.Context, outgoingDocID uuid.UUID, deptIDs []uuid.UUID) error
	FindRouteByIncomingDocID(ctx context.Context, incomingDocID uuid.UUID) (*domain.RouteStep, error)
	FindNextStep(ctx context.Context, outgoingDocID uuid.UUID, currentOrder int) (*domain.RouteStep, error)
	AttachIncomingDoc(ctx context.Context, routeID, incomingDocID uuid.UUID) error
}
