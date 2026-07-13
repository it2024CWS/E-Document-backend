package rejecteddoc

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"time"

	"github.com/google/uuid"
)

// Service defines the business logic for the rejected-document report.
type Service interface {
	GetRejectedDocs(ctx context.Context, deptID *uuid.UUID, source string, start, end *time.Time, page, limit int) ([]domain.RejectedDocResponse, int, error)
}

type service struct {
	repo Repository
}

// NewService creates a new rejected-document service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetRejectedDocs(ctx context.Context, deptID *uuid.UUID, source string, start, end *time.Time, page, limit int) ([]domain.RejectedDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindRejected(ctx, deptID, source, start, end, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get rejected documents", err)
	}
	if docs == nil {
		docs = []domain.RejectedDocResponse{}
	}
	return docs, total, nil
}
