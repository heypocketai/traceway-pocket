//go:build !pgch

package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
)

type exceptionStackTraceRepository struct{}

func (e *exceptionStackTraceRepository) InsertAsync(ctx context.Context, lines []models.ExceptionStackTrace) error {
	if len(lines) == 0 {
		return nil
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO exception_stack_traces (id, project_id, trace_id, trace_type, exception_hash, stack_trace, recorded_at, attributes, app_version, server_name, is_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, est := range lines {
		attributesJSON := "{}"
		if len(est.Attributes) != 0 {
			if attributesBytes, err := json.Marshal(est.Attributes); err == nil {
				attributesJSON = string(attributesBytes)
			}
		}

		isMessage := 0
		if est.IsMessage {
			isMessage = 1
		}

		traceType := est.TraceType
		if traceType == "" {
			traceType = "endpoint"
		}

		traceIdStr := ""
		if est.TraceId != nil {
			traceIdStr = est.TraceId.String()
		}

		if _, err := stmt.ExecContext(ctx,
			est.Id.String(),
			est.ProjectId.String(),
			traceIdStr,
			traceType,
			est.ExceptionHash,
			est.StackTrace,
			est.RecordedAt.UTC().Format(time.RFC3339Nano),
			attributesJSON,
			est.AppVersion,
			est.ServerName,
			isMessage,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (e *exceptionStackTraceRepository) CountBetween(ctx context.Context, projectId uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM exception_stack_traces WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?",
		projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano)).Scan(&count)
	return count, err
}

func (e *exceptionStackTraceRepository) FindGrouped(ctx context.Context, projectId uuid.UUID, fromDate, toDate time.Time, page, pageSize int, orderBy string, search string, searchType string, includeArchived bool) ([]models.ExceptionGroup, int64, error) {
	offset := (page - 1) * pageSize

	sortDirection := "DESC"
	if strings.HasSuffix(orderBy, "_asc") {
		orderBy = strings.TrimSuffix(orderBy, "_asc")
		sortDirection = "ASC"
	}

	allowedOrderBy := map[string]bool{
		"last_seen":  true,
		"first_seen": true,
		"count":      true,
	}
	if !allowedOrderBy[orderBy] {
		orderBy = "count"
	}

	whereClause := "e.project_id = ? AND e.recorded_at >= ? AND e.recorded_at <= ?"
	args := []interface{}{projectId.String(), fromDate.UTC().Format(time.RFC3339Nano), toDate.UTC().Format(time.RFC3339Nano)}

	if search != "" {
		whereClause += " AND INSTR(LOWER(e.stack_trace), LOWER(?)) > 0"
		args = append(args, search)
	}

	if searchType == "issues" {
		whereClause += " AND e.is_message = 0"
	} else if searchType == "messages" {
		whereClause += " AND e.is_message = 1"
	}

	havingClause := ""
	if !includeArchived {
		havingClause = " HAVING max_archived_at IS NULL OR MAX(e.recorded_at) > max_archived_at"
	}

	archiveSubquery := `LEFT JOIN (
		SELECT exception_hash, MAX(archived_at) as archived_at
		FROM archived_exceptions
		WHERE project_id = ?
		GROUP BY exception_hash
	) a ON e.exception_hash = a.exception_hash`

	countQuery := `SELECT COUNT(*) FROM (
		SELECT e.exception_hash, MAX(a.archived_at) as max_archived_at
		FROM exception_stack_traces e
		` + archiveSubquery + `
		WHERE ` + whereClause + `
		GROUP BY e.exception_hash` + havingClause + `
	)`

	countArgs := append([]interface{}{projectId.String()}, args...)
	var count int64
	err := db.DB.QueryRowContext(ctx, countQuery, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	fullQuery := `SELECT e.exception_hash, e.stack_trace, MAX(e.recorded_at) as last_seen, MIN(e.recorded_at) as first_seen, COUNT(*) as count, MAX(a.archived_at) as max_archived_at
		FROM exception_stack_traces e
		` + archiveSubquery + `
		WHERE ` + whereClause + `
		GROUP BY e.exception_hash` + havingClause + `
		ORDER BY ` + orderBy + ` ` + sortDirection + ` LIMIT ? OFFSET ?`

	queryArgs := append([]interface{}{projectId.String()}, args...)
	queryArgs = append(queryArgs, pageSize, offset)
	rows, err := db.DB.QueryContext(ctx, fullQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []models.ExceptionGroup
	for rows.Next() {
		var g models.ExceptionGroup
		var lastSeenStr, firstSeenStr string
		var maxArchivedAt sql.NullString
		if err := rows.Scan(&g.ExceptionHash, &g.StackTrace, &lastSeenStr, &firstSeenStr, &g.Count, &maxArchivedAt); err != nil {
			return nil, 0, err
		}
		g.LastSeen, _ = time.Parse(time.RFC3339Nano, lastSeenStr)
		g.FirstSeen, _ = time.Parse(time.RFC3339Nano, firstSeenStr)
		groups = append(groups, g)
	}

	return groups, count, nil
}

func (e *exceptionStackTraceRepository) FindByHash(ctx context.Context, projectId uuid.UUID, exceptionHash string, page, pageSize int) (*models.ExceptionGroup, []models.ExceptionStackTrace, int64, error) {
	offset := (page - 1) * pageSize

	var group models.ExceptionGroup
	var lastSeenStr, firstSeenStr string
	err := db.DB.QueryRowContext(ctx,
		`SELECT exception_hash, stack_trace, MAX(recorded_at) as last_seen, MIN(recorded_at) as first_seen, COUNT(*) as count
		FROM exception_stack_traces
		WHERE project_id = ? AND exception_hash = ?
		GROUP BY exception_hash`,
		projectId.String(), exceptionHash).Scan(&group.ExceptionHash, &group.StackTrace, &lastSeenStr, &firstSeenStr, &group.Count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, 0, nil
		}
		return nil, nil, 0, err
	}
	group.LastSeen, _ = time.Parse(time.RFC3339Nano, lastSeenStr)
	group.FirstSeen, _ = time.Parse(time.RFC3339Nano, firstSeenStr)

	rows, err := db.DB.QueryContext(ctx,
		`SELECT id, project_id, trace_id, trace_type, exception_hash, stack_trace, recorded_at, attributes, app_version, server_name, is_message
		FROM exception_stack_traces
		WHERE project_id = ? AND exception_hash = ?
		ORDER BY recorded_at DESC LIMIT ? OFFSET ?`,
		projectId.String(), exceptionHash, pageSize, offset)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	var occurrences []models.ExceptionStackTrace
	for rows.Next() {
		o := scanExceptionStackTrace(rows)
		if o == nil {
			return nil, nil, 0, fmt.Errorf("failed to scan exception stack trace row")
		}
		occurrences = append(occurrences, *o)
	}

	return &group, occurrences, int64(group.Count), nil
}

func (e *exceptionStackTraceRepository) CountByHour(ctx context.Context, projectId uuid.UUID, start, end time.Time) ([]models.TimeSeriesPoint, error) {
	query := `SELECT
		strftime('%Y-%m-%d %H:00:00', recorded_at) as hour,
		CAST(COUNT(*) AS REAL) as count
	FROM exception_stack_traces
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY hour
	ORDER BY hour ASC`

	rows, err := db.DB.QueryContext(ctx, query, projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.TimeSeriesPoint
	for rows.Next() {
		var p models.TimeSeriesPoint
		var tsStr string
		if err := rows.Scan(&tsStr, &p.Value); err != nil {
			return nil, err
		}
		p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr)
		points = append(points, p)
	}

	return points, nil
}

func (e *exceptionStackTraceRepository) CountByInterval(ctx context.Context, projectId uuid.UUID, start, end time.Time, intervalMinutes int) ([]models.TimeSeriesPoint, error) {
	intervalSeconds := intervalMinutes * 60

	query := `SELECT
		datetime((strftime('%s', recorded_at) / ?) * ?, 'unixepoch') as bucket,
		CAST(COUNT(*) AS REAL) as count
	FROM exception_stack_traces
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY bucket
	ORDER BY bucket ASC`

	rows, err := db.DB.QueryContext(ctx, query, intervalSeconds, intervalSeconds, projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []models.TimeSeriesPoint
	for rows.Next() {
		var p models.TimeSeriesPoint
		var tsStr string
		if err := rows.Scan(&tsStr, &p.Value); err != nil {
			return nil, err
		}
		p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr)
		points = append(points, p)
	}

	return points, nil
}

func (e *exceptionStackTraceRepository) GetHourlyTrendForHashes(ctx context.Context, projectId uuid.UUID, hashes []string, start, end time.Time) (map[string][]models.ExceptionTrendPoint, error) {
	if len(hashes) == 0 {
		return make(map[string][]models.ExceptionTrendPoint), nil
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(hashes)), ",")

	query := `SELECT
		exception_hash,
		strftime('%Y-%m-%d %H:00:00', recorded_at) as hour,
		COUNT(*) as count
	FROM exception_stack_traces
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ? AND exception_hash IN (` + placeholders + `)
	GROUP BY exception_hash, hour
	ORDER BY exception_hash, hour ASC`

	args := []interface{}{projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano)}
	for _, h := range hashes {
		args = append(args, h)
	}

	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]models.ExceptionTrendPoint)
	for rows.Next() {
		var hash string
		var point models.ExceptionTrendPoint
		var tsStr string
		if err := rows.Scan(&hash, &tsStr, &point.Count); err != nil {
			return nil, err
		}
		point.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr)
		result[hash] = append(result[hash], point)
	}

	return result, nil
}

func (e *exceptionStackTraceRepository) ArchiveByHashes(ctx context.Context, projectId uuid.UUID, hashes []string) error {
	if len(hashes) == 0 {
		return nil
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT OR REPLACE INTO archived_exceptions (project_id, exception_hash, archived_at) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, hash := range hashes {
		if _, err := stmt.ExecContext(ctx, projectId.String(), hash, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (e *exceptionStackTraceRepository) UnarchiveByHashes(ctx context.Context, projectId uuid.UUID, hashes []string) error {
	if len(hashes) == 0 {
		return nil
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(hashes)), ",")
	query := "DELETE FROM archived_exceptions WHERE project_id = ? AND exception_hash IN (" + placeholders + ")"

	args := []interface{}{projectId.String()}
	for _, h := range hashes {
		args = append(args, h)
	}

	_, err := db.DB.ExecContext(ctx, query, args...)
	return err
}

func (e *exceptionStackTraceRepository) IsArchived(ctx context.Context, projectId uuid.UUID, hash string) (bool, error) {
	var count int64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM archived_exceptions WHERE project_id = ? AND exception_hash = ?",
		projectId.String(), hash).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (e *exceptionStackTraceRepository) FindExceptionByTraceId(ctx context.Context, projectId uuid.UUID, traceId uuid.UUID) (*models.ExceptionStackTrace, error) {
	row := db.DB.QueryRowContext(ctx,
		`SELECT id, project_id, trace_id, trace_type, exception_hash, stack_trace, recorded_at, attributes, app_version, server_name, is_message
		FROM exception_stack_traces
		WHERE project_id = ? AND trace_id = ? AND is_message = 0
		LIMIT 1`,
		projectId.String(), traceId.String())

	est := scanExceptionStackTraceRow(row)
	return est, nil
}

func (e *exceptionStackTraceRepository) FindAllByTraceId(ctx context.Context, projectId uuid.UUID, traceId uuid.UUID) ([]models.ExceptionStackTrace, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT id, project_id, trace_id, trace_type, exception_hash, stack_trace, recorded_at, attributes, app_version, server_name, is_message
		FROM exception_stack_traces
		WHERE project_id = ? AND trace_id = ?
		ORDER BY recorded_at ASC`,
		projectId.String(), traceId.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ExceptionStackTrace
	for rows.Next() {
		o := scanExceptionStackTrace(rows)
		if o == nil {
			return nil, fmt.Errorf("failed to scan exception stack trace row")
		}
		results = append(results, *o)
	}

	return results, nil
}

func (e *exceptionStackTraceRepository) FindById(ctx context.Context, projectId uuid.UUID, id uuid.UUID) (*models.ExceptionStackTrace, error) {
	row := db.DB.QueryRowContext(ctx,
		`SELECT id, project_id, trace_id, trace_type, exception_hash, stack_trace, recorded_at, attributes, app_version, server_name, is_message
		FROM exception_stack_traces
		WHERE project_id = ? AND id = ?
		LIMIT 1`,
		projectId.String(), id.String())

	est := scanExceptionStackTraceRow(row)
	if est == nil {
		return nil, nil
	}
	return est, nil
}

func scanExceptionStackTraceRow(row *sql.Row) *models.ExceptionStackTrace {
	var est models.ExceptionStackTrace
	var idStr, projectIdStr, traceIdStr string
	var recordedAtStr, attributesJSON string
	var isMessage int

	err := row.Scan(&idStr, &projectIdStr, &traceIdStr, &est.TraceType, &est.ExceptionHash, &est.StackTrace,
		&recordedAtStr, &attributesJSON, &est.AppVersion, &est.ServerName, &isMessage)
	if err != nil {
		return nil
	}

	est.Id, _ = uuid.Parse(idStr)
	est.ProjectId, _ = uuid.Parse(projectIdStr)
	if traceIdStr != "" {
		parsed, err := uuid.Parse(traceIdStr)
		if err == nil {
			est.TraceId = &parsed
		}
	}
	est.RecordedAt, _ = time.Parse(time.RFC3339Nano, recordedAtStr)
	est.IsMessage = isMessage == 1

	if attributesJSON != "" && attributesJSON != "{}" {
		if err := json.Unmarshal([]byte(attributesJSON), &est.Attributes); err != nil {
			est.Attributes = nil
		}
	}

	return &est
}

func scanExceptionStackTrace(rows *sql.Rows) *models.ExceptionStackTrace {
	var est models.ExceptionStackTrace
	var idStr, projectIdStr, traceIdStr string
	var recordedAtStr, attributesJSON string
	var isMessage int

	err := rows.Scan(&idStr, &projectIdStr, &traceIdStr, &est.TraceType, &est.ExceptionHash, &est.StackTrace,
		&recordedAtStr, &attributesJSON, &est.AppVersion, &est.ServerName, &isMessage)
	if err != nil {
		return nil
	}

	est.Id, _ = uuid.Parse(idStr)
	est.ProjectId, _ = uuid.Parse(projectIdStr)
	if traceIdStr != "" {
		parsed, err := uuid.Parse(traceIdStr)
		if err == nil {
			est.TraceId = &parsed
		}
	}
	est.RecordedAt, _ = time.Parse(time.RFC3339Nano, recordedAtStr)
	est.IsMessage = isMessage == 1

	if attributesJSON != "" && attributesJSON != "{}" {
		if err := json.Unmarshal([]byte(attributesJSON), &est.Attributes); err != nil {
			est.Attributes = nil
		}
	}

	return &est
}

var ExceptionStackTraceRepository = exceptionStackTraceRepository{}
