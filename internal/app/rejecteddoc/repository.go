package rejecteddoc

import (
	"context"
	"e-document-backend/internal/domain"
	"time"

	"github.com/google/uuid"
)

// Repository defines data access for the rejected-document report.
type Repository interface {
	// FindRejected returns rejected documents from both sources (inbound incoming
	// docs and outbound outgoing docs), optionally filtered by department and a
	// rejection-date range, ordered newest first and paginated.
	FindRejected(ctx context.Context, deptID *uuid.UUID, start, end *time.Time, limit, offset int) ([]domain.RejectedDocResponse, int, error)
}
