// Package mailer is a thin SMTP transport used by the notification service.
// It is intentionally tiny: the caller renders subject + HTML; mailer only sends.
package mailer

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"e-document-backend/internal/config"

	"github.com/rs/zerolog/log"
)

// Message is what the notification service hands off to be transported.
type Message struct {
	To      []string
	Subject string
	HTML    string
}

// Mailer sends an email. Implementations must be safe for concurrent use.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

// New returns a real SMTP mailer if cfg.Host is set, otherwise a no-op mailer.
// This lets local dev run without SMTP credentials.
func New(cfg config.SMTPConfig) Mailer {
	if strings.TrimSpace(cfg.Host) == "" {
		log.Info().Msg("mailer: SMTP_HOST empty — using no-op mailer")
		return &noopMailer{}
	}
	return &smtpMailer{cfg: cfg}
}

type noopMailer struct{}

func (n *noopMailer) Send(_ context.Context, msg Message) error {
	log.Debug().Strs("to", msg.To).Str("subject", msg.Subject).Msg("noop mailer: dropping email")
	return nil
}

type smtpMailer struct {
	cfg config.SMTPConfig
}

func (s *smtpMailer) Send(_ context.Context, msg Message) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("mailer: no recipients")
	}

	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	from := s.cfg.FromEmail
	if from == "" {
		from = s.cfg.Username
	}
	fromHeader := from
	if s.cfg.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", s.cfg.FromName, from)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", fromHeader)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(msg.To, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", encodeSubject(msg.Subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.HTML)

	return smtp.SendMail(addr, auth, from, msg.To, []byte(b.String()))
}

// encodeSubject wraps non-ASCII subjects (e.g. Lao) in an RFC 2047 encoded-word
// so mail clients render them correctly.
func encodeSubject(s string) string {
	for _, r := range s {
		if r > 127 {
			return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
		}
	}
	return s
}
