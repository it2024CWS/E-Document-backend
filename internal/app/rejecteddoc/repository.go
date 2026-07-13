package rejecteddoc

import (
	"context"
	"e-document-backend/internal/domain"
	"time"

	"github.com/google/uuid"
)

// Repository defines data access for the rejected-document report.
type Repository interface {
	// FindRejected returns rejected documents from inbound and/or outbound legs,
	// optionally filtered by department, rejection-date range, and source
	// ("inbound" | "outbound" | ""), ordered newest first and paginated.
	FindRejected(ctx context.Context, deptID *uuid.UUID, source string, start, end *time.Time, limit, offset int) ([]domain.RejectedDocResponse, int, error)
}
