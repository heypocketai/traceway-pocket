//go:build !pgch

package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/traceway/backend/app/db"
)

type FiredNotification struct {
	ProjectId   uuid.UUID
	RuleId      int
	RuleType    string
	RuleName    string
	ChannelType string
	ChannelName string
	Severity    string
	Subject     string
	Body        string
	Status      string
	ErrorMsg    string
	Endpoint    string
	FiredAt     time.Time
}

type firedNotificationRepository struct{}

func (r *firedNotificationRepository) Insert(ctx context.Context, n FiredNotification) error {
	_, err := db.DB.ExecContext(ctx,
		`INSERT INTO fired_notifications (project_id, rule_id, rule_type, rule_name, channel_type, channel_name, severity, subject, body, status, error_message, endpoint, fired_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ProjectId.String(), n.RuleId, n.RuleType, n.RuleName,
		n.ChannelType, n.ChannelName, n.Severity, n.Subject, n.Body,
		n.Status, n.ErrorMsg, n.Endpoint, n.FiredAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

var FiredNotificationRepository = firedNotificationRepository{}
