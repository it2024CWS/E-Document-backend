// Package notification is the orchestrator for outbound email alerts.
// It owns the data lookups (owner email, department name, actor name) so that
// document services can call it with just IDs after their state transitions.
//
// All public methods swallow errors: notifications are best-effort side effects
// and must never block or fail an API response. Callers invoke via `go svc.X(...)`.
package notification

import (
	"context"
	"strings"
	"time"

	"e-document-backend/internal/app/department"
	"e-document-backend/internal/app/incomingdoc"
	"e-document-backend/internal/app/outgoingdoc"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/pkg/mailer"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Service composes lookups + template render + mailer transport.
type Service struct {
	mailer   mailer.Mailer
	users    user.Repository
	depts    department.Repository
	outgoing outgoingdoc.Repository
	incoming incomingdoc.Repository
}

// New constructs the notification service.
func New(
	m mailer.Mailer,
	users user.Repository,
	depts department.Repository,
	outgoing outgoingdoc.Repository,
	incoming incomingdoc.Repository,
) *Service {
	return &Service{
		mailer:   m,
		users:    users,
		depts:    depts,
		outgoing: outgoing,
		incoming: incoming,
	}
}

// IncomingReceived notifies the owner that their doc was received by a department.
func (s *Service) IncomingReceived(ctx context.Context, incomingDocID, actorID uuid.UUID) {
	s.notifyIncoming(ctx, eventIncomingReceived, incomingDocID, &actorID)
}

// IncomingApproved notifies the owner that their doc was approved by a department head.
func (s *Service) IncomingApproved(ctx context.Context, incomingDocID, actorID uuid.UUID) {
	s.notifyIncoming(ctx, eventIncomingApproved, incomingDocID, &actorID)
}

// IncomingRejected notifies the owner that their doc was rejected by a department head.
func (s *Service) IncomingRejected(ctx context.Context, incomingDocID, actorID uuid.UUID) {
	s.notifyIncoming(ctx, eventIncomingRejected, incomingDocID, &actorID)
}

// OutgoingOwnerApproved notifies the owner that their own dept head approved dispatch.
func (s *Service) OutgoingOwnerApproved(ctx context.Context, outgoingDocID uuid.UUID, actorID *uuid.UUID) {
	s.notifyOutgoing(ctx, eventOwnerApproved, outgoingDocID, actorID)
}

// OutgoingOwnerRejected notifies the owner that their own dept head rejected dispatch.
func (s *Service) OutgoingOwnerRejected(ctx context.Context, outgoingDocID uuid.UUID, actorID *uuid.UUID) {
	s.notifyOutgoing(ctx, eventOwnerRejected, outgoingDocID, actorID)
}

func (s *Service) notifyIncoming(ctx context.Context, kind eventKind, incomingDocID uuid.UUID, actorID *uuid.UUID) {
	doc, err := s.incoming.FindByID(ctx, incomingDocID)
	if err != nil || doc == nil {
		log.Warn().Err(err).Str("incoming_doc_id", incomingDocID.String()).Msg("notification: failed to load incoming doc")
		return
	}
	if doc.OutgoingDocID == nil {
		// Standalone (legacy) incoming doc with no owner to notify.
		return
	}
	outDoc, err := s.outgoing.FindByID(ctx, *doc.OutgoingDocID)
	if err != nil || outDoc == nil {
		log.Warn().Err(err).Str("outgoing_doc_id", doc.OutgoingDocID.String()).Msg("notification: failed to load outgoing doc")
		return
	}

	deptName := doc.DeptName
	if deptName == "" && doc.DeptID != nil {
		deptName = s.deptNameByID(ctx, *doc.DeptID)
	}

	data := templateData{
		DocNo:     firstNonEmpty(doc.DocNo, outDoc.DocNo, shortID(outDoc.ID)),
		DocName:   firstNonEmpty(doc.DocName, outDoc.DocName),
		DeptName:  deptName,
		Remark:    doc.Remark,
		ActorName: s.userDisplayName(ctx, actorID),
		Timestamp: time.Now().Format("2006-01-02 15:04"),
	}
	s.sendToOwner(ctx, kind, outDoc.CreatedBy, data)
}

func (s *Service) notifyOutgoing(ctx context.Context, kind eventKind, outgoingDocID uuid.UUID, actorID *uuid.UUID) {
	outDoc, err := s.outgoing.FindByID(ctx, outgoingDocID)
	if err != nil || outDoc == nil {
		log.Warn().Err(err).Str("outgoing_doc_id", outgoingDocID.String()).Msg("notification: failed to load outgoing doc")
		return
	}

	deptName := outDoc.OwnerDeptName
	if deptName == "" && outDoc.OwnerDeptID != nil {
		deptName = s.deptNameByID(ctx, *outDoc.OwnerDeptID)
	}

	data := templateData{
		DocNo:     firstNonEmpty(outDoc.DocNo, shortID(outDoc.ID)),
		DocName:   outDoc.DocName,
		DeptName:  deptName,
		ActorName: s.userDisplayName(ctx, actorID),
		Timestamp: time.Now().Format("2006-01-02 15:04"),
	}
	s.sendToOwner(ctx, kind, outDoc.CreatedBy, data)
}

func (s *Service) sendToOwner(ctx context.Context, kind eventKind, ownerID *uuid.UUID, data templateData) {
	if ownerID == nil {
		return
	}
	owner, err := s.users.FindByID(ctx, ownerID.String())
	if err != nil || owner == nil {
		log.Warn().Err(err).Str("owner_id", ownerID.String()).Msg("notification: failed to load owner user")
		return
	}
	to := strings.TrimSpace(owner.Email)
	if to == "" {
		return
	}
	data.OwnerName = userDisplay(owner)

	subject, html, err := render(kind, data)
	if err != nil {
		log.Error().Err(err).Msg("notification: failed to render template")
		return
	}
	if err := s.mailer.Send(ctx, mailer.Message{To: []string{to}, Subject: subject, HTML: html}); err != nil {
		log.Error().Err(err).Str("to", to).Msg("notification: failed to send email")
		return
	}
	log.Info().Str("to", to).Str("subject", subject).Msg("notification: email sent")
}

func (s *Service) deptNameByID(ctx context.Context, id uuid.UUID) string {
	d, err := s.depts.FindByID(ctx, id.String())
	if err != nil || d == nil {
		return ""
	}
	return d.DeptName
}

func (s *Service) userDisplayName(ctx context.Context, id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	u, err := s.users.FindByID(ctx, id.String())
	if err != nil || u == nil {
		return ""
	}
	return userDisplay(u)
}

func userDisplay(u *domain.User) string {
	full := strings.TrimSpace(u.Firstname + " " + u.Lastname)
	if full != "" {
		return full
	}
	if u.Nickname != "" {
		return u.Nickname
	}
	return u.Username
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func shortID(id uuid.UUID) string {
	s := id.String()
	if len(s) >= 8 {
		return s[:8]
	}
	return s
}
