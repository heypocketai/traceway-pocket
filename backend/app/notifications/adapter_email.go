package notifications

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/tracewayapp/traceway/backend/app/config"
	"github.com/tracewayapp/traceway/backend/app/services"
)

type EmailAdapter struct {
	Recipients []string `json:"recipients"`
}

func (a *EmailAdapter) Type() string { return "email" }

func (a *EmailAdapter) Validate() error {
	if len(a.Recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if len(a.Recipients) > 10 {
		return fmt.Errorf("maximum 10 recipients allowed")
	}
	for _, r := range a.Recipients {
		if !strings.Contains(r, "@") {
			return fmt.Errorf("invalid email address: %s", r)
		}
	}
	return nil
}

func (a *EmailAdapter) Send(ctx context.Context, msg Message) error {
	emailSvc := services.EmailService
	if emailSvc == nil {
		return fmt.Errorf("email service not initialized")
	}

	prefix := ""
	switch msg.Severity {
	case SeverityCritical:
		prefix = "[CRITICAL] "
	case SeverityWarning:
		prefix = "[WARNING] "
	case SeverityInfo:
		prefix = "[INFO] "
	}

	subject := prefix + msg.Subject

	if !emailSvc.IsEnabled() {
		config.Logf("[EMAIL LOG] To: %s\nSubject: %s\nBody:\n%s", strings.Join(a.Recipients, ", "), subject, msg.Body)
		return nil
	}

	cfg := config.Config
	from := cfg.SMTPFrom
	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)

	emailMsg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, strings.Join(a.Recipients, ", "), subject, msg.Body)

	return sendMailWithTimeout(ctx, addr, auth, from, a.Recipients, []byte(emailMsg))
}

func sendMailWithTimeout(ctx context.Context, addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("SMTP dial failed: %w", err)
	}

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("SMTP client failed: %w", err)
	}
	defer client.Close()

	conn.SetDeadline(time.Now().Add(10 * time.Second))

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("SMTP STARTTLS failed: %w", err)
		}
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL failed: %w", err)
	}
	for _, r := range to {
		if err := client.Rcpt(r); err != nil {
			return fmt.Errorf("SMTP RCPT failed: %w", err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("SMTP write failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("SMTP close data failed: %w", err)
	}
	client.Quit()
	return nil
}
