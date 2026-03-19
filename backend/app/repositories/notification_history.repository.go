package repositories

import (
	"database/sql"
	"time"

	"github.com/tracewayapp/traceway/backend/app/models"

	"github.com/google/uuid"
	"github.com/tracewayapp/lit/v2"
)

type notificationHistoryRepository struct{}

func (r *notificationHistoryRepository) FindByProject(tx *sql.Tx, projectId uuid.UUID, page, pageSize int, search string, fromDate, toDate *time.Time) ([]*models.NotificationHistory, int64, error) {
	params := lit.P{"project_id": projectId}
	whereClause := "WHERE project_id = :project_id"

	if search != "" {
		whereClause += " AND (rule_name ILIKE :search OR channel_name ILIKE :search OR subject ILIKE :search)"
		params["search"] = "%" + search + "%"
	}

	if fromDate != nil {
		whereClause += " AND created_at >= :from_date"
		params["from_date"] = *fromDate
	}
	if toDate != nil {
		whereClause += " AND created_at <= :to_date"
		params["to_date"] = *toDate
	}

	countResult, err := lit.SelectSingleNamed[models.CountResult](
		tx,
		"SELECT COUNT(*) as count FROM notification_history "+whereClause,
		params,
	)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if countResult != nil {
		total = int64(countResult.Count)
	}

	offset := (page - 1) * pageSize
	params["limit"] = pageSize
	params["offset"] = offset
	items, err := lit.SelectNamed[models.NotificationHistory](
		tx,
		"SELECT id, project_id, rule_id, channel_id, rule_type, rule_name, channel_name, severity, subject, body, status, error_message, url, created_at FROM notification_history "+whereClause+" ORDER BY created_at DESC LIMIT :limit OFFSET :offset",
		params,
	)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *notificationHistoryRepository) Create(tx *sql.Tx, history *models.NotificationHistory) (int, error) {
	return lit.Insert[models.NotificationHistory](tx, history)
}

var NotificationHistoryRepository = notificationHistoryRepository{}
