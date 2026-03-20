package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/hooks"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/repositories"
	traceway "go.tracewayapp.com"
)

type newErrorConfig struct {
	IgnorePatterns []string `json:"ignorePatterns"`
}

func registerReportHook() {
	hooks.RegisterReportHook(func(event hooks.ReportEvent) {
		if len(event.ExceptionHashes) == 0 {
			return
		}
		go evaluateEventRules(event)
	})
}

func evaluateEventRules(event hooks.ReportEvent) {
	rules, err := db.ExecuteTransaction(func(tx *sql.Tx) ([]*models.NotificationRuleWithChannel, error) {
		return repositories.NotificationRuleRepository.FindEnabledEventRules(tx, event.ProjectId)
	})
	if err != nil {
		traceway.CaptureException(fmt.Errorf("failed to load event notification rules: %w", err))
		return
	}

	ctx := context.Background()

	for _, rule := range rules {
		if rule.SnoozedUntil != nil && rule.SnoozedUntil.After(time.Now()) {
			continue
		}

		switch rule.RuleType {
		case "new_error":
			evaluateNewError(ctx, rule, event)
		case "error_regression":
			evaluateErrorRegression(ctx, rule, event)
		}
	}
}

func extractErrorType(stackTrace string) string {
	if stackTrace == "" {
		return "Unknown Error"
	}
	lines := strings.SplitN(stackTrace, "\n", 2)
	if len(lines) > 0 {
		line := strings.TrimSpace(lines[0])
		if idx := strings.Index(line, ":"); idx > 0 {
			return line[:idx]
		}
		return line
	}
	return "Unknown Error"
}

func getProjectName(projectId uuid.UUID) string {
	project, err := db.ExecuteTransaction(func(tx *sql.Tx) (*models.Project, error) {
		return repositories.ProjectRepository.FindById(tx, projectId)
	})
	if err != nil || project == nil {
		return ""
	}
	return project.Name
}

func shouldIgnore(errorType string, patterns []string) bool {
	lower := strings.ToLower(errorType)
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		pattern = strings.ReplaceAll(pattern, "*", "")
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
