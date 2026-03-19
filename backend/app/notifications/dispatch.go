package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	traceway "go.tracewayapp.com"
)

func sanitizeForDB(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "\uFFFD")
	}
	return s
}

func dispatch(rule *models.NotificationRuleWithChannel, msg Message) {
	channel, dbErr := db.ExecuteTransaction(func(tx *sql.Tx) (*models.NotificationChannel, error) {
		return repositories.NotificationChannelRepository.FindById(tx, rule.ChannelId)
	})
	if dbErr != nil || channel == nil {
		recordHistory(rule, msg, "failed", "failed to load channel")
		recordFiredNotification(rule, msg, "failed", "failed to load channel")
		return
	}

	adapter, err := NewAdapter(channel.ChannelType, channel.Config)
	if err != nil {
		log.Printf("[notif] dispatch: ruleId=%d adapter error: %v", rule.Id, err)
		recordHistory(rule, msg, "failed", err.Error())
		recordFiredNotification(rule, msg, "failed", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msg.RuleType = rule.RuleType
	msg.RuleName = rule.Name
	if rule.Severity != "" {
		msg.Severity = Severity(rule.Severity)
	}

	err = adapter.Send(ctx, msg)
	if err != nil {
		log.Printf("[notif] dispatch: ruleId=%d send FAILED: %v", rule.Id, err)
		recordHistory(rule, msg, "failed", err.Error())
		recordFiredNotification(rule, msg, "failed", err.Error())
		traceway.CaptureException(fmt.Errorf("notification dispatch failed (rule=%d, channel=%s): %w", rule.Id, rule.ChannelName, err))
		return
	}

	log.Printf("[notif] dispatch: ruleId=%d sent successfully via %s", rule.Id, channel.ChannelType)
	recordHistory(rule, msg, "sent", "")
	recordFiredNotification(rule, msg, "sent", "")
	cooldowns.recordFire(rule.Id)
}

func recordFiredNotification(rule *models.NotificationRuleWithChannel, msg Message, status string, errorMsg string) {
	go func() {
		err := repositories.FiredNotificationRepository.Insert(context.Background(), repositories.FiredNotification{
			ProjectId:   rule.ProjectId,
			RuleId:      rule.Id,
			RuleType:    rule.RuleType,
			RuleName:    rule.Name,
			ChannelType: rule.ChannelType,
			ChannelName: rule.ChannelName,
			Severity:    string(msg.Severity),
			Subject:     sanitizeForDB(msg.Subject),
			Body:        sanitizeForDB(msg.Body),
			Status:      status,
			ErrorMsg:    sanitizeForDB(errorMsg),
			Endpoint:    msg.Endpoint,
			FiredAt:     time.Now().UTC(),
		})
		if err != nil {
			traceway.CaptureException(fmt.Errorf("failed to record fired notification to ClickHouse: %w", err))
		}
	}()
}

func recordHistory(rule *models.NotificationRuleWithChannel, msg Message, status string, errorMsg string) {
	var errMsgPtr *string
	if errorMsg != "" {
		errMsgPtr = &errorMsg
	}

	ruleId := rule.Id
	channelId := rule.ChannelId

	history := &models.NotificationHistory{
		ProjectId:    rule.ProjectId,
		RuleId:       &ruleId,
		ChannelId:    &channelId,
		RuleType:     rule.RuleType,
		RuleName:     rule.Name,
		ChannelName:  rule.ChannelName,
		Severity:     string(msg.Severity),
		Subject:      sanitizeForDB(msg.Subject),
		Body:         sanitizeForDB(msg.Body),
		Status:       status,
		ErrorMessage: errMsgPtr,
		URL:          msg.URL,
		CreatedAt:    time.Now().UTC(),
	}

	if errMsgPtr != nil {
		sanitized := sanitizeForDB(*errMsgPtr)
		errMsgPtr = &sanitized
		history.ErrorMessage = errMsgPtr
	}

	_, dbErr := db.ExecuteTransaction(func(tx *sql.Tx) (int, error) {
		return repositories.NotificationHistoryRepository.Create(tx, history)
	})
	if dbErr != nil {
		log.Printf("[notif] recordHistory FAILED: ruleId=%d subject=%q err=%v", rule.Id, msg.Subject, dbErr)
		traceway.CaptureException(fmt.Errorf("failed to record notification history: %w", dbErr))
	}
}
