//go:build !pgch

package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
)

type taskRepository struct{}

func (e *taskRepository) InsertAsync(ctx context.Context, lines []models.Task) error {
	if len(lines) == 0 {
		return nil
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO tasks (id, project_id, task_name, duration, recorded_at, client_ip, attributes, app_version, server_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range lines {
		attributesJSON := "{}"
		if len(t.Attributes) != 0 {
			if attributesBytes, err := json.Marshal(t.Attributes); err == nil {
				attributesJSON = string(attributesBytes)
			}
		}
		if _, err := stmt.ExecContext(ctx,
			t.Id.String(), t.ProjectId.String(), t.TaskName,
			int64(t.Duration), t.RecordedAt.UTC().Format(time.RFC3339Nano),
			t.ClientIP, attributesJSON, t.AppVersion, t.ServerName,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (e *taskRepository) CountBetween(ctx context.Context, projectId uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM tasks WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?",
		projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano),
	).Scan(&count)
	return count, err
}

func (e *taskRepository) FindAll(ctx context.Context, projectId uuid.UUID, fromDate, toDate time.Time, page, pageSize int, orderBy string) ([]models.Task, int64, error) {
	var count int64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM tasks WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?",
		projectId.String(), fromDate.UTC().Format(time.RFC3339Nano), toDate.UTC().Format(time.RFC3339Nano),
	).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	allowedOrderBy := map[string]bool{
		"recorded_at": true,
		"duration":    true,
	}
	if !allowedOrderBy[orderBy] {
		orderBy = "recorded_at"
	}

	query := "SELECT id, project_id, task_name, duration, recorded_at, client_ip, attributes, app_version, server_name FROM tasks WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY " + orderBy + " DESC LIMIT ? OFFSET ?"
	rows, err := db.DB.QueryContext(ctx, query,
		projectId.String(), fromDate.UTC().Format(time.RFC3339Nano), toDate.UTC().Format(time.RFC3339Nano),
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}

	return tasks, count, nil
}

func (e *taskRepository) FindGroupedByTaskName(ctx context.Context, projectId uuid.UUID, fromDate, toDate time.Time, page, pageSize int, orderBy string, sortDirection string) ([]models.TaskStats, int64, error) {
	fromStr := fromDate.UTC().Format(time.RFC3339Nano)
	toStr := toDate.UTC().Format(time.RFC3339Nano)
	pidStr := projectId.String()

	var totalCount int64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(DISTINCT task_name) FROM tasks WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?",
		pidStr, fromStr, toStr,
	).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	needsGoSort := orderBy == "p50_duration" || orderBy == "p95_duration" || orderBy == "impact"

	sortDir := "DESC"
	if sortDirection == "asc" {
		sortDir = "ASC"
	}

	offset := (page - 1) * pageSize

	var baseQuery string
	var baseArgs []interface{}

	if needsGoSort {
		baseQuery = `SELECT task_name, COUNT(*) as count, AVG(duration) as avg_duration, MAX(recorded_at) as last_seen
			FROM tasks
			WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
			GROUP BY task_name`
		baseArgs = []interface{}{pidStr, fromStr, toStr}
	} else {
		orderExpr := map[string]string{
			"count":     "count",
			"last_seen": "last_seen",
		}
		expr, ok := orderExpr[orderBy]
		if !ok {
			expr = "count"
		}
		baseQuery = `SELECT task_name, COUNT(*) as count, AVG(duration) as avg_duration, MAX(recorded_at) as last_seen
			FROM tasks
			WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
			GROUP BY task_name
			ORDER BY ` + expr + ` ` + sortDir + `
			LIMIT ? OFFSET ?`
		baseArgs = []interface{}{pidStr, fromStr, toStr, pageSize, offset}
	}

	rows, err := db.DB.QueryContext(ctx, baseQuery, baseArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	type intermediate struct {
		taskName    string
		count       uint64
		avgDuration float64
		lastSeen    string
	}

	var intermediates []intermediate
	for rows.Next() {
		var it intermediate
		if err := rows.Scan(&it.taskName, &it.count, &it.avgDuration, &it.lastSeen); err != nil {
			return nil, 0, err
		}
		intermediates = append(intermediates, it)
	}

	var stats []models.TaskStats
	for _, it := range intermediates {
		durRows, err := db.DB.QueryContext(ctx,
			"SELECT duration FROM tasks WHERE project_id = ? AND task_name = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY duration ASC",
			pidStr, it.taskName, fromStr, toStr,
		)
		if err != nil {
			return nil, 0, err
		}

		var sorted []float64
		for durRows.Next() {
			var d float64
			if err := durRows.Scan(&d); err != nil {
				durRows.Close()
				return nil, 0, err
			}
			sorted = append(sorted, d)
		}
		durRows.Close()

		ls, _ := time.Parse(time.RFC3339Nano, it.lastSeen)
		stats = append(stats, models.TaskStats{
			TaskName:    it.taskName,
			Count:       it.count,
			P50Duration: time.Duration(computePercentile(sorted, 0.5)),
			P95Duration: time.Duration(computePercentile(sorted, 0.95)),
			AvgDuration: time.Duration(it.avgDuration),
			LastSeen:    ls,
		})
	}

	if needsGoSort {
		switch orderBy {
		case "p50_duration":
			sort.Slice(stats, func(i, j int) bool {
				if sortDir == "ASC" {
					return stats[i].P50Duration < stats[j].P50Duration
				}
				return stats[i].P50Duration > stats[j].P50Duration
			})
		case "p95_duration":
			sort.Slice(stats, func(i, j int) bool {
				if sortDir == "ASC" {
					return stats[i].P95Duration < stats[j].P95Duration
				}
				return stats[i].P95Duration > stats[j].P95Duration
			})
		case "impact":
			sort.Slice(stats, func(i, j int) bool {
				impactI := float64(stats[i].Count) * float64(stats[i].P95Duration-stats[i].P50Duration)
				impactJ := float64(stats[j].Count) * float64(stats[j].P95Duration-stats[j].P50Duration)
				if sortDir == "ASC" {
					return impactI < impactJ
				}
				return impactI > impactJ
			})
		}

		end := offset + pageSize
		if end > len(stats) {
			end = len(stats)
		}
		if offset > len(stats) {
			stats = nil
		} else {
			stats = stats[offset:end]
		}
	}

	return stats, totalCount, nil
}

func (e *taskRepository) FindByTaskName(ctx context.Context, projectId uuid.UUID, taskName string, fromDate, toDate time.Time, page, pageSize int, orderBy string, sortDirection string) ([]models.Task, int64, error) {
	fromStr := fromDate.UTC().Format(time.RFC3339Nano)
	toStr := toDate.UTC().Format(time.RFC3339Nano)
	pidStr := projectId.String()

	var count int64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM tasks WHERE project_id = ? AND task_name = ? AND recorded_at >= ? AND recorded_at <= ?",
		pidStr, taskName, fromStr, toStr,
	).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	allowedOrderBy := map[string]bool{
		"recorded_at": true,
		"duration":    true,
	}
	if !allowedOrderBy[orderBy] {
		orderBy = "recorded_at"
	}

	sortDir := "DESC"
	if sortDirection == "asc" {
		sortDir = "ASC"
	}

	query := "SELECT id, project_id, task_name, duration, recorded_at, client_ip, attributes, app_version, server_name FROM tasks WHERE project_id = ? AND task_name = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY " + orderBy + " " + sortDir + " LIMIT ? OFFSET ?"
	rows, err := db.DB.QueryContext(ctx, query, pidStr, taskName, fromStr, toStr, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}

	return tasks, count, nil
}

func (e *taskRepository) FindById(ctx context.Context, projectId, taskId uuid.UUID) (*models.Task, error) {
	row := db.DB.QueryRowContext(ctx,
		`SELECT id, project_id, task_name, duration, recorded_at, client_ip, attributes, app_version, server_name
		FROM tasks
		WHERE project_id = ? AND id = ?
		LIMIT 1`,
		projectId.String(), taskId.String(),
	)

	var t models.Task
	var idStr, projectIdStr string
	var recordedAtStr string
	var dur int64
	var attributesJSON string

	err := row.Scan(&idStr, &projectIdStr, &t.TaskName, &dur, &recordedAtStr, &t.ClientIP, &attributesJSON, &t.AppVersion, &t.ServerName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	t.Id, _ = uuid.Parse(idStr)
	t.ProjectId, _ = uuid.Parse(projectIdStr)
	t.Duration = time.Duration(dur)
	t.RecordedAt, _ = time.Parse(time.RFC3339Nano, recordedAtStr)

	if attributesJSON != "" && attributesJSON != "{}" {
		if err := json.Unmarshal([]byte(attributesJSON), &t.Attributes); err != nil {
			t.Attributes = nil
		}
	}

	return &t, nil
}

func (e *taskRepository) CountByHour(ctx context.Context, projectId uuid.UUID, start, end time.Time) ([]models.TimeSeriesPoint, error) {
	query := `SELECT
		strftime('%Y-%m-%d %H:00:00', recorded_at) as hour,
		CAST(COUNT(*) AS REAL) as count
	FROM tasks
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY hour
	ORDER BY hour ASC`

	rows, err := db.DB.QueryContext(ctx, query,
		projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.TimeSeriesPoint
	for rows.Next() {
		var tsStr string
		var p models.TimeSeriesPoint
		if err := rows.Scan(&tsStr, &p.Value); err != nil {
			return nil, err
		}
		p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr)
		points = append(points, p)
	}

	return points, nil
}

func (e *taskRepository) AvgDurationByHour(ctx context.Context, projectId uuid.UUID, start, end time.Time) ([]models.TimeSeriesPoint, error) {
	query := `SELECT
		strftime('%Y-%m-%d %H:00:00', recorded_at) as hour,
		AVG(duration) / 1000000.0 as avg_duration_ms
	FROM tasks
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY hour
	ORDER BY hour ASC`

	rows, err := db.DB.QueryContext(ctx, query,
		projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.TimeSeriesPoint
	for rows.Next() {
		var tsStr string
		var p models.TimeSeriesPoint
		if err := rows.Scan(&tsStr, &p.Value); err != nil {
			return nil, err
		}
		p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr)
		points = append(points, p)
	}

	return points, nil
}

func (e *taskRepository) CountByInterval(ctx context.Context, projectId uuid.UUID, start, end time.Time, intervalMinutes int) ([]models.TimeSeriesPoint, error) {
	query := fmt.Sprintf(`SELECT
		datetime((strftime('%%s', recorded_at) / (%d * 60)) * (%d * 60), 'unixepoch') as bucket,
		CAST(COUNT(*) AS REAL) as count
	FROM tasks
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY bucket
	ORDER BY bucket ASC`, intervalMinutes, intervalMinutes)

	rows, err := db.DB.QueryContext(ctx, query,
		projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.TimeSeriesPoint
	for rows.Next() {
		var tsStr string
		var p models.TimeSeriesPoint
		if err := rows.Scan(&tsStr, &p.Value); err != nil {
			return nil, err
		}
		p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr)
		points = append(points, p)
	}

	return points, nil
}

func (e *taskRepository) AvgDurationByInterval(ctx context.Context, projectId uuid.UUID, start, end time.Time, intervalMinutes int) ([]models.TimeSeriesPoint, error) {
	query := fmt.Sprintf(`SELECT
		datetime((strftime('%%s', recorded_at) / (%d * 60)) * (%d * 60), 'unixepoch') as bucket,
		AVG(duration) / 1000000.0 as avg_duration_ms
	FROM tasks
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY bucket
	ORDER BY bucket ASC`, intervalMinutes, intervalMinutes)

	rows, err := db.DB.QueryContext(ctx, query,
		projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.TimeSeriesPoint
	for rows.Next() {
		var tsStr string
		var p models.TimeSeriesPoint
		if err := rows.Scan(&tsStr, &p.Value); err != nil {
			return nil, err
		}
		p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr)
		points = append(points, p)
	}

	return points, nil
}

func (e *taskRepository) FindWorstTasks(ctx context.Context, projectId uuid.UUID, start, end time.Time, limit int) ([]models.TaskStats, error) {
	fromStr := start.UTC().Format(time.RFC3339Nano)
	toStr := end.UTC().Format(time.RFC3339Nano)
	pidStr := projectId.String()

	rows, err := db.DB.QueryContext(ctx,
		`SELECT task_name, COUNT(*) as count, AVG(duration) as avg_duration, MAX(recorded_at) as last_seen
		FROM tasks
		WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
		GROUP BY task_name`,
		pidStr, fromStr, toStr,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type intermediate struct {
		taskName    string
		count       uint64
		avgDuration float64
		lastSeen    string
	}

	var intermediates []intermediate
	for rows.Next() {
		var it intermediate
		if err := rows.Scan(&it.taskName, &it.count, &it.avgDuration, &it.lastSeen); err != nil {
			return nil, err
		}
		intermediates = append(intermediates, it)
	}

	var stats []models.TaskStats
	for _, it := range intermediates {
		durRows, err := db.DB.QueryContext(ctx,
			"SELECT duration FROM tasks WHERE project_id = ? AND task_name = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY duration ASC",
			pidStr, it.taskName, fromStr, toStr,
		)
		if err != nil {
			return nil, err
		}

		var sorted []float64
		for durRows.Next() {
			var d float64
			if err := durRows.Scan(&d); err != nil {
				durRows.Close()
				return nil, err
			}
			sorted = append(sorted, d)
		}
		durRows.Close()

		ls, _ := time.Parse(time.RFC3339Nano, it.lastSeen)
		stats = append(stats, models.TaskStats{
			TaskName:    it.taskName,
			Count:       it.count,
			P50Duration: time.Duration(computePercentile(sorted, 0.5)),
			P95Duration: time.Duration(computePercentile(sorted, 0.95)),
			AvgDuration: time.Duration(it.avgDuration),
			LastSeen:    ls,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		impactI := float64(stats[i].Count) * float64(stats[i].P95Duration-stats[i].P50Duration)
		impactJ := float64(stats[j].Count) * float64(stats[j].P95Duration-stats[j].P50Duration)
		return impactI > impactJ
	})

	if limit > len(stats) {
		limit = len(stats)
	}
	return stats[:limit], nil
}

func (e *taskRepository) GetTaskStats(ctx context.Context, projectId uuid.UUID, taskName string, start, end time.Time) (*models.TaskDetailStats, error) {
	fromStr := start.UTC().Format(time.RFC3339Nano)
	toStr := end.UTC().Format(time.RFC3339Nano)
	pidStr := projectId.String()

	durationMinutes := end.Sub(start).Minutes()
	if durationMinutes < 1 {
		durationMinutes = 1
	}

	var count int64
	var avgDur float64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*), AVG(duration) / 1000000.0 FROM tasks WHERE project_id = ? AND task_name = ? AND recorded_at >= ? AND recorded_at <= ?",
		pidStr, taskName, fromStr, toStr,
	).Scan(&count, &avgDur)
	if err != nil {
		return nil, err
	}

	durRows, err := db.DB.QueryContext(ctx,
		"SELECT duration FROM tasks WHERE project_id = ? AND task_name = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY duration ASC",
		pidStr, taskName, fromStr, toStr,
	)
	if err != nil {
		return nil, err
	}
	defer durRows.Close()

	var sorted []float64
	for durRows.Next() {
		var d float64
		if err := durRows.Scan(&d); err != nil {
			return nil, err
		}
		sorted = append(sorted, d)
	}

	nsToMs := 1000000.0

	return &models.TaskDetailStats{
		Count:          count,
		AvgDuration:    avgDur,
		MedianDuration: computePercentile(sorted, 0.5) / nsToMs,
		P95Duration:    computePercentile(sorted, 0.95) / nsToMs,
		P99Duration:    computePercentile(sorted, 0.99) / nsToMs,
		Throughput:     float64(count) / durationMinutes,
	}, nil
}

func scanTask(rows *sql.Rows) (models.Task, error) {
	var t models.Task
	var idStr, projectIdStr string
	var recordedAtStr string
	var dur int64
	var attributesJSON string

	if err := rows.Scan(&idStr, &projectIdStr, &t.TaskName, &dur, &recordedAtStr, &t.ClientIP, &attributesJSON, &t.AppVersion, &t.ServerName); err != nil {
		return t, err
	}

	t.Id, _ = uuid.Parse(idStr)
	t.ProjectId, _ = uuid.Parse(projectIdStr)
	t.Duration = time.Duration(dur)
	t.RecordedAt, _ = time.Parse(time.RFC3339Nano, recordedAtStr)

	if attributesJSON != "" && attributesJSON != "{}" {
		if err := json.Unmarshal([]byte(attributesJSON), &t.Attributes); err != nil {
			t.Attributes = nil
		}
	}

	return t, nil
}

var TaskRepository = taskRepository{}
