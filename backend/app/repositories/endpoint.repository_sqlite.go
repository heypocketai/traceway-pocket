//go:build !pgch

package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
)

type endpointRepository struct{}

func (e *endpointRepository) InsertAsync(ctx context.Context, lines []models.Endpoint) error {
	if len(lines) == 0 {
		return nil
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO endpoints (id, project_id, endpoint, duration, recorded_at, status_code, body_size, client_ip, attributes, app_version, server_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range lines {
		attributesJSON := "{}"
		if len(t.Attributes) != 0 {
			if b, err := json.Marshal(t.Attributes); err == nil {
				attributesJSON = string(b)
			}
		}
		if _, err := stmt.ExecContext(ctx,
			t.Id.String(), t.ProjectId.String(), t.Endpoint,
			int64(t.Duration), t.RecordedAt.UTC().Format(time.RFC3339Nano),
			t.StatusCode, t.BodySize, t.ClientIP, attributesJSON,
			t.AppVersion, t.ServerName,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (e *endpointRepository) CountBetween(ctx context.Context, projectId uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM endpoints WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?",
		projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano)).Scan(&count)
	return count, err
}

func (e *endpointRepository) FindAll(ctx context.Context, projectId uuid.UUID, fromDate, toDate time.Time, page, pageSize int, orderBy string) ([]models.Endpoint, int64, error) {
	var count int64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM endpoints WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?",
		projectId.String(), fromDate.UTC().Format(time.RFC3339Nano), toDate.UTC().Format(time.RFC3339Nano)).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	allowedOrderBy := map[string]bool{
		"recorded_at": true,
		"duration":    true,
		"status_code": true,
		"body_size":   true,
	}
	if !allowedOrderBy[orderBy] {
		orderBy = "recorded_at"
	}

	query := `SELECT id, project_id, endpoint, duration, recorded_at, status_code, body_size, client_ip, attributes, app_version, server_name
		FROM endpoints
		WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
		ORDER BY ` + orderBy + ` DESC LIMIT ? OFFSET ?`

	rows, err := db.DB.QueryContext(ctx, query,
		projectId.String(), fromDate.UTC().Format(time.RFC3339Nano), toDate.UTC().Format(time.RFC3339Nano),
		pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	endpoints, err := scanEndpoints(rows)
	if err != nil {
		return nil, 0, err
	}

	return endpoints, count, nil
}

func (e *endpointRepository) FindGroupedByEndpoint(ctx context.Context, projectId uuid.UUID, fromDate, toDate time.Time, page, pageSize int, orderBy string, sortDirection string, search string) ([]models.EndpointStats, int64, error) {
	fromStr := fromDate.UTC().Format(time.RFC3339Nano)
	toStr := toDate.UTC().Format(time.RFC3339Nano)
	pidStr := projectId.String()

	whereClause := "e.project_id = ? AND e.recorded_at >= ? AND e.recorded_at <= ?"
	args := []interface{}{pidStr, fromStr, toStr}
	if search != "" {
		whereClause += " AND INSTR(LOWER(e.endpoint), LOWER(?)) > 0"
		args = append(args, search)
	}

	var totalEndpoints int64
	countQuery := "SELECT COUNT(DISTINCT e.endpoint) FROM endpoints e WHERE " + whereClause
	if err := db.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalEndpoints); err != nil {
		return nil, 0, err
	}

	groupQuery := `SELECT
		e.endpoint,
		COUNT(*) as total_count,
		AVG(e.duration) as avg_duration,
		MAX(e.recorded_at) as last_seen,
		COALESCE(s.offset_ms, 0) as offset_ms,
		SUM(CASE WHEN e.status_code >= 500 THEN 1 ELSE 0 END) as server_error_count,
		SUM(CASE WHEN e.status_code >= 400 AND e.status_code < 500 THEN 1 ELSE 0 END) as client_error_count,
		SUM(CASE WHEN e.duration <= (750000000 + COALESCE(s.offset_ms, 0) * 1000000) AND e.status_code < 500 THEN 1 ELSE 0 END) as satisfied_count,
		SUM(CASE WHEN e.duration > (750000000 + COALESCE(s.offset_ms, 0) * 1000000) AND e.duration <= (1500000000 + COALESCE(s.offset_ms, 0) * 1000000) AND e.status_code < 500 THEN 1 ELSE 0 END) as tolerating_count,
		SUM(CASE WHEN e.duration > (1500000000 + COALESCE(s.offset_ms, 0) * 1000000) OR e.status_code >= 500 THEN 1 ELSE 0 END) as bad_count
	FROM endpoints e
	LEFT JOIN slow_endpoints s ON e.endpoint = s.endpoint AND e.project_id = s.project_id
	WHERE ` + whereClause + `
	GROUP BY e.endpoint`

	rows, err := db.DB.QueryContext(ctx, groupQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	type groupRow struct {
		endpoint         string
		totalCount       uint64
		avgDuration      float64
		lastSeen         string
		offsetMs         uint32
		serverErrorCount uint64
		clientErrorCount uint64
		satisfiedCount   uint64
		toleratingCount  uint64
		badCount         uint64
	}

	var groups []groupRow
	for rows.Next() {
		var g groupRow
		if err := rows.Scan(&g.endpoint, &g.totalCount, &g.avgDuration, &g.lastSeen,
			&g.offsetMs, &g.serverErrorCount, &g.clientErrorCount,
			&g.satisfiedCount, &g.toleratingCount, &g.badCount); err != nil {
			return nil, 0, err
		}
		groups = append(groups, g)
	}

	var stats []models.EndpointStats
	for _, g := range groups {
		durations, err := fetchSortedDurations(ctx, pidStr, g.endpoint, fromStr, toStr)
		if err != nil {
			return nil, 0, err
		}

		p50 := computePercentile(durations, 0.5)
		p95 := computePercentile(durations, 0.95)
		p99 := computePercentile(durations, 0.99)

		impact := computeImpactScore(g.endpoint, g.totalCount, g.satisfiedCount, g.toleratingCount, g.badCount, g.clientErrorCount, p99, g.offsetMs)

		lastSeen, _ := time.Parse(time.RFC3339Nano, g.lastSeen)

		s := models.EndpointStats{
			Endpoint:    g.endpoint,
			Count:       g.totalCount,
			P50Duration: time.Duration(p50),
			P95Duration: time.Duration(p95),
			P99Duration: time.Duration(p99),
			AvgDuration: time.Duration(g.avgDuration),
			LastSeen:    lastSeen,
			Impact:      impact,
			ImpactReason: ComputeImpactReason(g.endpoint, g.totalCount, g.satisfiedCount, g.toleratingCount,
				g.badCount, g.clientErrorCount, p99, g.offsetMs),
		}
		stats = append(stats, s)
	}

	sortEndpointStats(stats, orderBy, sortDirection)

	start := (page - 1) * pageSize
	if start > len(stats) {
		start = len(stats)
	}
	endIdx := start + pageSize
	if endIdx > len(stats) {
		endIdx = len(stats)
	}
	page_stats := stats[start:endIdx]

	return page_stats, totalEndpoints, nil
}

func (e *endpointRepository) FindByEndpoint(ctx context.Context, projectId uuid.UUID, endpoint string, fromDate, toDate time.Time, page, pageSize int, orderBy string, sortDirection string) ([]models.Endpoint, int64, error) {
	pidStr := projectId.String()
	fromStr := fromDate.UTC().Format(time.RFC3339Nano)
	toStr := toDate.UTC().Format(time.RFC3339Nano)

	var count int64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM endpoints WHERE project_id = ? AND endpoint = ? AND recorded_at >= ? AND recorded_at <= ?",
		pidStr, endpoint, fromStr, toStr).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	allowedOrderBy := map[string]bool{
		"recorded_at": true,
		"duration":    true,
		"status_code": true,
		"body_size":   true,
	}
	if !allowedOrderBy[orderBy] {
		orderBy = "recorded_at"
	}

	sortDir := "DESC"
	if sortDirection == "asc" {
		sortDir = "ASC"
	}

	query := `SELECT id, project_id, endpoint, duration, recorded_at, status_code, body_size, client_ip, attributes, app_version, server_name
		FROM endpoints
		WHERE project_id = ? AND endpoint = ? AND recorded_at >= ? AND recorded_at <= ?
		ORDER BY ` + orderBy + ` ` + sortDir + ` LIMIT ? OFFSET ?`

	rows, err := db.DB.QueryContext(ctx, query, pidStr, endpoint, fromStr, toStr, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	endpoints, err := scanEndpoints(rows)
	if err != nil {
		return nil, 0, err
	}

	return endpoints, count, nil
}

func (e *endpointRepository) FindById(ctx context.Context, projectId, endpointId uuid.UUID) (*models.Endpoint, error) {
	query := `SELECT id, project_id, endpoint, duration, recorded_at, status_code, body_size, client_ip, attributes, app_version, server_name
		FROM endpoints
		WHERE project_id = ? AND id = ?
		LIMIT 1`

	var t models.Endpoint
	var idStr, projectIdStr, recordedAtStr, attributesJSON string
	err := db.DB.QueryRowContext(ctx, query, projectId.String(), endpointId.String()).Scan(
		&idStr, &projectIdStr, &t.Endpoint, &t.Duration, &recordedAtStr,
		&t.StatusCode, &t.BodySize, &t.ClientIP, &attributesJSON, &t.AppVersion, &t.ServerName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	t.Id, _ = uuid.Parse(idStr)
	t.ProjectId, _ = uuid.Parse(projectIdStr)
	t.RecordedAt, _ = time.Parse(time.RFC3339Nano, recordedAtStr)
	if attributesJSON != "" && attributesJSON != "{}" {
		if err := json.Unmarshal([]byte(attributesJSON), &t.Attributes); err != nil {
			t.Attributes = nil
		}
	}

	return &t, nil
}

func (e *endpointRepository) CountByHour(ctx context.Context, projectId uuid.UUID, start, end time.Time) ([]models.TimeSeriesPoint, error) {
	query := `SELECT
		strftime('%Y-%m-%d %H:00:00', recorded_at) as hour,
		CAST(COUNT(*) AS REAL) as count
	FROM endpoints
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY hour
	ORDER BY hour ASC`

	return queryTimeSeries(ctx, query, projectId.String(), start, end)
}

func (e *endpointRepository) AvgDurationByHour(ctx context.Context, projectId uuid.UUID, start, end time.Time) ([]models.TimeSeriesPoint, error) {
	query := `SELECT
		strftime('%Y-%m-%d %H:00:00', recorded_at) as hour,
		AVG(duration) / 1000000.0 as avg_duration_ms
	FROM endpoints
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY hour
	ORDER BY hour ASC`

	return queryTimeSeries(ctx, query, projectId.String(), start, end)
}

func (e *endpointRepository) ErrorRateByHour(ctx context.Context, projectId uuid.UUID, start, end time.Time) ([]models.TimeSeriesPoint, error) {
	query := `SELECT
		strftime('%Y-%m-%d %H:00:00', recorded_at) as hour,
		SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as error_rate
	FROM endpoints
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY hour
	ORDER BY hour ASC`

	return queryTimeSeries(ctx, query, projectId.String(), start, end)
}

func (e *endpointRepository) CountByInterval(ctx context.Context, projectId uuid.UUID, start, end time.Time, intervalMinutes int) ([]models.TimeSeriesPoint, error) {
	secs := intervalMinutes * 60
	query := fmt.Sprintf(`SELECT
		datetime((strftime('%%s', recorded_at) / %d) * %d, 'unixepoch') as bucket,
		CAST(COUNT(*) AS REAL) as count
	FROM endpoints
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY bucket
	ORDER BY bucket ASC`, secs, secs)

	return queryTimeSeries(ctx, query, projectId.String(), start, end)
}

func (e *endpointRepository) AvgDurationByInterval(ctx context.Context, projectId uuid.UUID, start, end time.Time, intervalMinutes int) ([]models.TimeSeriesPoint, error) {
	secs := intervalMinutes * 60
	query := fmt.Sprintf(`SELECT
		datetime((strftime('%%s', recorded_at) / %d) * %d, 'unixepoch') as bucket,
		AVG(duration) / 1000000.0 as avg_duration_ms
	FROM endpoints
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY bucket
	ORDER BY bucket ASC`, secs, secs)

	return queryTimeSeries(ctx, query, projectId.String(), start, end)
}

func (e *endpointRepository) ErrorRateByInterval(ctx context.Context, projectId uuid.UUID, start, end time.Time, intervalMinutes int) ([]models.TimeSeriesPoint, error) {
	secs := intervalMinutes * 60
	query := fmt.Sprintf(`SELECT
		datetime((strftime('%%s', recorded_at) / %d) * %d, 'unixepoch') as bucket,
		SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as error_rate
	FROM endpoints
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY bucket
	ORDER BY bucket ASC`, secs, secs)

	return queryTimeSeries(ctx, query, projectId.String(), start, end)
}

func (e *endpointRepository) FindWorstEndpoints(ctx context.Context, projectId uuid.UUID, start, end time.Time, limit int) ([]models.EndpointStats, error) {
	fromStr := start.UTC().Format(time.RFC3339Nano)
	toStr := end.UTC().Format(time.RFC3339Nano)
	pidStr := projectId.String()

	groupQuery := `SELECT
		e.endpoint,
		COUNT(*) as total_count,
		AVG(e.duration) as avg_duration,
		MAX(e.recorded_at) as last_seen,
		COALESCE(s.offset_ms, 0) as offset_ms,
		SUM(CASE WHEN e.status_code >= 500 THEN 1 ELSE 0 END) as server_error_count,
		SUM(CASE WHEN e.status_code >= 400 AND e.status_code < 500 THEN 1 ELSE 0 END) as client_error_count,
		SUM(CASE WHEN e.duration <= (750000000 + COALESCE(s.offset_ms, 0) * 1000000) AND e.status_code < 500 THEN 1 ELSE 0 END) as satisfied_count,
		SUM(CASE WHEN e.duration > (750000000 + COALESCE(s.offset_ms, 0) * 1000000) AND e.duration <= (1500000000 + COALESCE(s.offset_ms, 0) * 1000000) AND e.status_code < 500 THEN 1 ELSE 0 END) as tolerating_count,
		SUM(CASE WHEN e.duration > (1500000000 + COALESCE(s.offset_ms, 0) * 1000000) OR e.status_code >= 500 THEN 1 ELSE 0 END) as bad_count
	FROM endpoints e
	LEFT JOIN slow_endpoints s ON e.endpoint = s.endpoint AND e.project_id = s.project_id
	WHERE e.project_id = ? AND e.recorded_at >= ? AND e.recorded_at <= ?
	GROUP BY e.endpoint`

	rows, err := db.DB.QueryContext(ctx, groupQuery, pidStr, fromStr, toStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type groupRow struct {
		endpoint         string
		totalCount       uint64
		avgDuration      float64
		lastSeen         string
		offsetMs         uint32
		serverErrorCount uint64
		clientErrorCount uint64
		satisfiedCount   uint64
		toleratingCount  uint64
		badCount         uint64
	}

	var groups []groupRow
	for rows.Next() {
		var g groupRow
		if err := rows.Scan(&g.endpoint, &g.totalCount, &g.avgDuration, &g.lastSeen,
			&g.offsetMs, &g.serverErrorCount, &g.clientErrorCount,
			&g.satisfiedCount, &g.toleratingCount, &g.badCount); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	var stats []models.EndpointStats
	for _, g := range groups {
		durations, err := fetchSortedDurations(ctx, pidStr, g.endpoint, fromStr, toStr)
		if err != nil {
			return nil, err
		}

		p50 := computePercentile(durations, 0.5)
		p95 := computePercentile(durations, 0.95)
		p99 := computePercentile(durations, 0.99)

		impact := computeImpactScore(g.endpoint, g.totalCount, g.satisfiedCount, g.toleratingCount, g.badCount, g.clientErrorCount, p99, g.offsetMs)

		lastSeen, _ := time.Parse(time.RFC3339Nano, g.lastSeen)

		s := models.EndpointStats{
			Endpoint:    g.endpoint,
			Count:       g.totalCount,
			P50Duration: time.Duration(p50),
			P95Duration: time.Duration(p95),
			P99Duration: time.Duration(p99),
			AvgDuration: time.Duration(g.avgDuration),
			LastSeen:    lastSeen,
			Impact:      impact,
			ImpactReason: ComputeImpactReason(g.endpoint, g.totalCount, g.satisfiedCount, g.toleratingCount,
				g.badCount, g.clientErrorCount, p99, g.offsetMs),
		}
		stats = append(stats, s)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Impact > stats[j].Impact
	})

	if limit > 0 && limit < len(stats) {
		stats = stats[:limit]
	}

	return stats, nil
}

func (e *endpointRepository) GetEndpointStats(ctx context.Context, projectId uuid.UUID, endpoint string, start, end time.Time) (*models.EndpointDetailStats, error) {
	pidStr := projectId.String()
	fromStr := start.UTC().Format(time.RFC3339Nano)
	toStr := end.UTC().Format(time.RFC3339Nano)

	durationMinutes := end.Sub(start).Minutes()
	if durationMinutes < 1 {
		durationMinutes = 1
	}

	query := `SELECT
		COUNT(*) as count,
		CASE WHEN COUNT(*) > 0 THEN AVG(duration) / 1000000.0 ELSE 0 END as avg_duration_ms,
		CASE WHEN COUNT(*) > 0 THEN SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) ELSE 0 END as error_rate,
		SUM(CASE WHEN duration <= 500000000 AND status_code < 500 THEN 1 ELSE 0 END) +
			SUM(CASE WHEN duration > 500000000 AND duration <= 2000000000 AND status_code < 500 THEN 1 ELSE 0 END) * 0.5 as satisfied_tolerating
	FROM endpoints
	WHERE project_id = ? AND endpoint = ? AND recorded_at >= ? AND recorded_at <= ?`

	var stats models.EndpointDetailStats
	var count int64
	var satisfiedTolerating float64

	err := db.DB.QueryRowContext(ctx, query, pidStr, endpoint, fromStr, toStr).Scan(
		&count, &stats.AvgDuration, &stats.ErrorRate, &satisfiedTolerating)
	if err != nil {
		return nil, err
	}

	stats.Count = count
	if count > 0 {
		stats.Apdex = satisfiedTolerating / float64(count)
	}
	stats.Throughput = float64(count) / durationMinutes

	durations, err := fetchSortedDurations(ctx, pidStr, endpoint, fromStr, toStr)
	if err != nil {
		return nil, err
	}

	if len(durations) > 0 {
		stats.MedianDuration = computePercentile(durations, 0.5) / 1000000.0
		stats.P95Duration = computePercentile(durations, 0.95) / 1000000.0
		stats.P99Duration = computePercentile(durations, 0.99) / 1000000.0
	}

	return &stats, nil
}

func (e *endpointRepository) GetEndpointStackedChart(ctx context.Context, projectId uuid.UUID, start, end time.Time, intervalMinutes int, metricType string) (*models.EndpointStackedChartResponse, error) {
	pidStr := projectId.String()
	fromStr := start.UTC().Format(time.RFC3339Nano)
	toStr := end.UTC().Format(time.RFC3339Nano)

	// Step 1: Get top 5 endpoints by metric
	topEndpoints, err := getTopEndpointsByMetric(ctx, pidStr, fromStr, toStr, metricType)
	if err != nil {
		return nil, err
	}

	if len(topEndpoints) == 0 {
		return &models.EndpointStackedChartResponse{
			Endpoints: []string{},
			Series:    []models.EndpointTimeSeriesPoint{},
		}, nil
	}

	// Step 2: Build CASE expression for categorization
	caseExpr := "CASE "
	caseArgs := make([]interface{}, 0, len(topEndpoints))
	for _, ep := range topEndpoints {
		caseExpr += "WHEN endpoint = ? THEN ? "
		caseArgs = append(caseArgs, ep, ep)
	}
	caseExpr += "ELSE 'Other' END"

	var metricExpr string
	switch metricType {
	case "total_time":
		metricExpr = "COUNT(*) * AVG(duration) / 1000000.0"
	case "p95":
		metricExpr = "AVG(duration) / 1000000.0"
	case "p99":
		metricExpr = "AVG(duration) / 1000000.0"
	default:
		metricExpr = "AVG(duration) / 1000000.0"
	}

	secs := intervalMinutes * 60
	timeSeriesQuery := fmt.Sprintf(`SELECT
		datetime((strftime('%%s', recorded_at) / %d) * %d, 'unixepoch') as bucket,
		%s as endpoint_category,
		%s as metric_value
	FROM endpoints
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	GROUP BY bucket, endpoint_category
	ORDER BY bucket ASC, endpoint_category ASC`, secs, secs, caseExpr, metricExpr)

	args := make([]interface{}, 0, len(caseArgs)+3)
	args = append(args, caseArgs...)
	args = append(args, pidStr, fromStr, toStr)

	// For p50/p95/p99 we need per-bucket per-category percentile computation
	if metricType == "p50" || metricType == "p95" || metricType == "p99" || metricType == "" {
		return e.getStackedChartWithPercentiles(ctx, pidStr, fromStr, toStr, secs, topEndpoints, metricType)
	}

	rows, err := db.DB.QueryContext(ctx, timeSeriesQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var series []models.EndpointTimeSeriesPoint
	for rows.Next() {
		var p models.EndpointTimeSeriesPoint
		var tsStr string
		if err := rows.Scan(&tsStr, &p.Endpoint, &p.Value); err != nil {
			return nil, err
		}
		p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr)
		series = append(series, p)
	}

	endpointSet := make(map[string]bool)
	for _, p := range series {
		endpointSet[p.Endpoint] = true
	}

	finalEndpoints := make([]string, 0, len(topEndpoints)+1)
	finalEndpoints = append(finalEndpoints, topEndpoints...)
	if endpointSet["Other"] {
		finalEndpoints = append(finalEndpoints, "Other")
	}

	return &models.EndpointStackedChartResponse{
		Endpoints: finalEndpoints,
		Series:    series,
	}, nil
}

func (e *endpointRepository) getStackedChartWithPercentiles(ctx context.Context, pidStr, fromStr, toStr string, bucketSecs int, topEndpoints []string, metricType string) (*models.EndpointStackedChartResponse, error) {
	percentile := 0.5
	switch metricType {
	case "p95":
		percentile = 0.95
	case "p99":
		percentile = 0.99
	}

	topSet := make(map[string]bool, len(topEndpoints))
	for _, ep := range topEndpoints {
		topSet[ep] = true
	}

	query := fmt.Sprintf(`SELECT
		datetime((strftime('%%s', recorded_at) / %d) * %d, 'unixepoch') as bucket,
		endpoint,
		duration
	FROM endpoints
	WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
	ORDER BY bucket ASC, endpoint ASC, duration ASC`, bucketSecs, bucketSecs)

	rows, err := db.DB.QueryContext(ctx, query, pidStr, fromStr, toStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type bucketKey struct {
		bucket   string
		category string
	}
	durationMap := make(map[bucketKey][]float64)

	for rows.Next() {
		var bucketStr, endpoint string
		var duration float64
		if err := rows.Scan(&bucketStr, &endpoint, &duration); err != nil {
			return nil, err
		}

		category := "Other"
		if topSet[endpoint] {
			category = endpoint
		}

		key := bucketKey{bucket: bucketStr, category: category}
		durationMap[key] = append(durationMap[key], duration)
	}

	var series []models.EndpointTimeSeriesPoint
	for key, durations := range durationMap {
		sort.Float64s(durations)
		val := computePercentile(durations, percentile) / 1000000.0

		ts, _ := time.Parse("2006-01-02 15:04:05", key.bucket)
		series = append(series, models.EndpointTimeSeriesPoint{
			Timestamp: ts,
			Endpoint:  key.category,
			Value:     val,
		})
	}

	sort.Slice(series, func(i, j int) bool {
		if series[i].Timestamp.Equal(series[j].Timestamp) {
			return series[i].Endpoint < series[j].Endpoint
		}
		return series[i].Timestamp.Before(series[j].Timestamp)
	})

	endpointSet := make(map[string]bool)
	for _, p := range series {
		endpointSet[p.Endpoint] = true
	}

	finalEndpoints := make([]string, 0, len(topEndpoints)+1)
	finalEndpoints = append(finalEndpoints, topEndpoints...)
	if endpointSet["Other"] {
		finalEndpoints = append(finalEndpoints, "Other")
	}

	return &models.EndpointStackedChartResponse{
		Endpoints: finalEndpoints,
		Series:    series,
	}, nil
}

func (e *endpointRepository) GetSlowEndpoint(ctx context.Context, projectId uuid.UUID, endpoint string) (uint32, string, error) {
	var offsetMs uint32
	var reason string
	err := db.DB.QueryRowContext(ctx,
		"SELECT offset_ms, reason FROM slow_endpoints WHERE project_id = ? AND endpoint = ?",
		projectId.String(), endpoint).Scan(&offsetMs, &reason)
	return offsetMs, reason, err
}

func (e *endpointRepository) UpsertSlowEndpoint(ctx context.Context, projectId uuid.UUID, endpoint string, offsetMs uint32, reason string) error {
	_, err := db.DB.ExecContext(ctx,
		"INSERT OR REPLACE INTO slow_endpoints (project_id, endpoint, offset_ms, reason) VALUES (?, ?, ?, ?)",
		projectId.String(), endpoint, offsetMs, reason)
	return err
}

// --- helpers ---

func scanEndpoints(rows *sql.Rows) ([]models.Endpoint, error) {
	var endpoints []models.Endpoint
	for rows.Next() {
		var t models.Endpoint
		var idStr, projectIdStr, recordedAtStr, attributesJSON string
		if err := rows.Scan(&idStr, &projectIdStr, &t.Endpoint, &t.Duration, &recordedAtStr,
			&t.StatusCode, &t.BodySize, &t.ClientIP, &attributesJSON, &t.AppVersion, &t.ServerName); err != nil {
			return nil, err
		}
		t.Id, _ = uuid.Parse(idStr)
		t.ProjectId, _ = uuid.Parse(projectIdStr)
		t.RecordedAt, _ = time.Parse(time.RFC3339Nano, recordedAtStr)
		if attributesJSON != "" && attributesJSON != "{}" {
			if err := json.Unmarshal([]byte(attributesJSON), &t.Attributes); err != nil {
				t.Attributes = nil
			}
		}
		endpoints = append(endpoints, t)
	}
	return endpoints, nil
}

func fetchSortedDurations(ctx context.Context, pidStr, endpoint, fromStr, toStr string) ([]float64, error) {
	rows, err := db.DB.QueryContext(ctx,
		"SELECT duration FROM endpoints WHERE project_id = ? AND endpoint = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY duration ASC",
		pidStr, endpoint, fromStr, toStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var durations []float64
	for rows.Next() {
		var d float64
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		durations = append(durations, d)
	}
	return durations, nil
}

func computeImpactScore(endpoint string, total, satisfiedCount, toleratingCount, badCount, clientErrorCount uint64, p99Ns float64, offsetMs uint32) float64 {
	if total == 0 {
		return 0
	}

	totalF := float64(total)
	badF := float64(badCount)
	clientF := float64(clientErrorCount)
	offsetNs := float64(offsetMs) * 1_000_000

	apdexScore := 1.0 - (float64(satisfiedCount)+float64(toleratingCount)*0.5)/totalF

	badRate := badF / totalF
	var errorRateScore float64
	switch {
	case badRate > 0.33:
		errorRateScore = 0.75
	case badRate > 0.20:
		errorRateScore = 0.50
	case badRate > 0.10:
		errorRateScore = 0.25
	}

	adjustedP99 := p99Ns - offsetNs
	var p99Score float64
	switch {
	case adjustedP99 > 8_000_000_000:
		p99Score = 0.75
	case adjustedP99 > 6_000_000_000:
		p99Score = 0.50
	case adjustedP99 > 3_000_000_000:
		p99Score = 0.25
	}

	var clientErrorScore float64
	if endpoint != "UNMATCHED" && total > 10 {
		clientRate := clientF / totalF
		switch {
		case clientRate > 0.50:
			clientErrorScore = 0.75
		case clientRate > 0.25:
			clientErrorScore = 0.50
		}
	}

	var volumeScore float64
	switch {
	case badRate > 0.10 && badCount >= 500:
		volumeScore = 0.75
	case badRate > 0.10 && badCount >= 50:
		volumeScore = 0.50
	case badRate > 0.05 && badCount >= 2000:
		volumeScore = 0.75
	case badRate > 0.05 && badCount >= 500:
		volumeScore = 0.50
	case badRate > 0.05 && badCount >= 50:
		volumeScore = 0.25
	case badRate > 0.01 && badCount >= 10000:
		volumeScore = 0.75
	case badRate > 0.01 && badCount >= 2000:
		volumeScore = 0.50
	case badRate > 0.01 && badCount >= 500:
		volumeScore = 0.25
	}

	return math.Max(apdexScore, math.Max(errorRateScore, math.Max(p99Score, math.Max(clientErrorScore, volumeScore))))
}

func sortEndpointStats(stats []models.EndpointStats, orderBy string, sortDirection string) {
	desc := sortDirection != "asc"

	sort.Slice(stats, func(i, j int) bool {
		var less bool
		switch orderBy {
		case "count":
			less = stats[i].Count < stats[j].Count
		case "p50_duration":
			less = stats[i].P50Duration < stats[j].P50Duration
		case "p95_duration":
			less = stats[i].P95Duration < stats[j].P95Duration
		case "p99_duration":
			less = stats[i].P99Duration < stats[j].P99Duration
		case "avg_duration":
			less = stats[i].AvgDuration < stats[j].AvgDuration
		case "last_seen":
			less = stats[i].LastSeen.Before(stats[j].LastSeen)
		default:
			less = stats[i].Impact < stats[j].Impact
		}
		if desc {
			return !less
		}
		return less
	})
}

func queryTimeSeries(ctx context.Context, query, pidStr string, start, end time.Time) ([]models.TimeSeriesPoint, error) {
	rows, err := db.DB.QueryContext(ctx, query, pidStr, start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano))
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

func getTopEndpointsByMetric(ctx context.Context, pidStr, fromStr, toStr, metricType string) ([]string, error) {
	if metricType == "total_time" {
		query := `SELECT endpoint, COUNT(*) * AVG(duration) / 1000000.0 as metric_value
			FROM endpoints
			WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
			GROUP BY endpoint
			ORDER BY metric_value DESC
			LIMIT 5`

		rows, err := db.DB.QueryContext(ctx, query, pidStr, fromStr, toStr)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var endpoints []string
		for rows.Next() {
			var ep string
			var val float64
			if err := rows.Scan(&ep, &val); err != nil {
				return nil, err
			}
			endpoints = append(endpoints, ep)
		}
		return endpoints, nil
	}

	// For p50/p95/p99: need to fetch durations per endpoint, compute percentile, sort
	percentile := 0.5
	switch metricType {
	case "p95":
		percentile = 0.95
	case "p99":
		percentile = 0.99
	}

	epQuery := `SELECT DISTINCT endpoint FROM endpoints WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?`
	rows, err := db.DB.QueryContext(ctx, epQuery, pidStr, fromStr, toStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allEndpoints []string
	for rows.Next() {
		var ep string
		if err := rows.Scan(&ep); err != nil {
			return nil, err
		}
		allEndpoints = append(allEndpoints, ep)
	}

	type epMetric struct {
		endpoint string
		value    float64
	}

	var metrics []epMetric
	for _, ep := range allEndpoints {
		durations, err := fetchSortedDurations(ctx, pidStr, ep, fromStr, toStr)
		if err != nil {
			return nil, err
		}
		val := computePercentile(durations, percentile) / 1000000.0
		metrics = append(metrics, epMetric{endpoint: ep, value: val})
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].value > metrics[j].value
	})

	limit := 5
	if limit > len(metrics) {
		limit = len(metrics)
	}

	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		result[i] = metrics[i].endpoint
	}

	return result, nil
}

var EndpointRepository = endpointRepository{}
