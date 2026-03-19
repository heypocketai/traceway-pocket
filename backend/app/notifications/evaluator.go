package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tracewayapp/traceway/backend/app/config"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	traceway "go.tracewayapp.com"
)

func StartEvaluator(ctx context.Context) {
	config.Logln("Starting notification evaluator")
	startDedupPurger(ctx)
	registerReportHook()
	go startPolledLoop(ctx)
}

func startPolledLoop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			evaluatePolledRules(ctx)
		}
	}
}

func evaluatePolledRules(ctx context.Context) {
	rules, err := db.ExecuteTransaction(func(tx *sql.Tx) ([]*models.NotificationRuleWithChannel, error) {
		return repositories.NotificationRuleRepository.FindEnabledPolledRules(tx)
	})
	if err != nil {
		traceway.CaptureException(fmt.Errorf("failed to load polled notification rules: %w", err))
		return
	}

	for _, rule := range rules {
		if rule.SnoozedUntil != nil && rule.SnoozedUntil.After(time.Now()) {
			continue
		}

		if !cooldowns.canFire(rule.Id, rule.CooldownMinutes) {
			continue
		}

		evaluator, ok := polledEvaluators[rule.RuleType]
		if !ok {
			continue
		}

		nr := &models.NotificationRule{
			Id:              rule.Id,
			ProjectId:       rule.ProjectId,
			ChannelId:       rule.ChannelId,
			Name:            rule.Name,
			RuleType:        rule.RuleType,
			Config:          rule.Config,
			Enabled:         rule.Enabled,
			CooldownMinutes: rule.CooldownMinutes,
			SnoozedUntil:    rule.SnoozedUntil,
			CreatedBy:       rule.CreatedBy,
			CreatedAt:       rule.CreatedAt,
			UpdatedAt:       rule.UpdatedAt,
		}
		result, err := evaluator(ctx, nr, rule.ProjectId)
		if err != nil {
			traceway.CaptureException(fmt.Errorf("notification evaluator error (rule=%d, type=%s): %w", rule.Id, rule.RuleType, err))
			continue
		}

		if result != nil && result.Fired {
			if len(result.Messages) > 0 {
				for _, msg := range result.Messages {
					dispatch(rule, msg)
				}
			} else {
				dispatch(rule, result.Message)
			}
		}
	}
}
