//go:build !pgch

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
)

type metricPointRepository struct{}

func (r *metricPointRepository) InsertAsync(ctx context.Context, points []models.MetricPoint) error {
	if len(points) == 0 {
		return nil
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO metric_points (project_id, name, value, tags, recorded_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range points {
		tags := p.Tags
		if tags == nil {
			tags = map[string]string{}
		}
		tagsJSON, err := json.Marshal(tags)
		if err != nil {
			return err
		}

		if _, err := stmt.ExecContext(ctx,
			p.ProjectId.String(),
			p.Name,
			p.Value,
			string(tagsJSON),
			p.RecordedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *metricPointRepository) QueryTimeSeries(ctx context.Context, projectId uuid.UUID, name string, from, to time.Time, intervalMinutes int, aggregation string, tagFilters map[string]string, groupBy string) (map[string][]models.TimeSeriesPoint, error) {
	secs := intervalMinutes * 60

	aggFunc := sqliteAggregationFunc(aggregation)

	query := fmt.Sprintf("SELECT datetime((strftime('%%s', recorded_at) / %d) * %d, 'unixepoch') AS bucket", secs, secs)

	args := []interface{}{}

	hasGroupBy := groupBy != ""
	if hasGroupBy {
		query += ", json_extract(tags, '$.' || ?) AS group_key"
		args = append(args, groupBy)
	}

	query += ", " + aggFunc + " AS agg_value FROM metric_points WHERE project_id = ? AND name = ? AND recorded_at >= ? AND recorded_at <= ?"
	args = append(args, projectId.String(), name, from.UTC().Format(time.RFC3339Nano), to.UTC().Format(time.RFC3339Nano))

	for k, v := range tagFilters {
		query += " AND json_extract(tags, '$.' || ?) = ?"
		args = append(args, k, v)
	}

	query += " GROUP BY bucket"
	if hasGroupBy {
		query += ", group_key"
	}
	query += " ORDER BY bucket ASC"

	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]models.TimeSeriesPoint)
	for rows.Next() {
		var bucketStr string
		var value float64
		groupKey := "__all__"

		if hasGroupBy {
			var groupKeyNullable *string
			if err := rows.Scan(&bucketStr, &groupKeyNullable, &value); err != nil {
				return nil, err
			}
			if groupKeyNullable != nil {
				groupKey = *groupKeyNullable
			}
		} else {
			if err := rows.Scan(&bucketStr, &value); err != nil {
				return nil, err
			}
		}

		bucket, err := time.Parse("2006-01-02 15:04:05", bucketStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bucket time %q: %w", bucketStr, err)
		}

		if groupKey == "" {
			groupKey = "(empty)"
		}
		result[groupKey] = append(result[groupKey], models.TimeSeriesPoint{
			Timestamp: bucket,
			Value:     value,
		})
	}
	return result, nil
}

func (r *metricPointRepository) DiscoverMetrics(ctx context.Context, projectId uuid.UUID, from, to time.Time) ([]models.DiscoveredMetric, error) {
	nameRows, err := db.DB.QueryContext(ctx,
		"SELECT DISTINCT name FROM metric_points WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY name ASC",
		projectId.String(), from.UTC().Format(time.RFC3339Nano), to.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer nameRows.Close()

	var names []string
	for nameRows.Next() {
		var name string
		if err := nameRows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	var metrics []models.DiscoveredMetric
	for _, name := range names {
		tagRows, err := db.DB.QueryContext(ctx,
			`SELECT DISTINCT j.key FROM metric_points, json_each(tags) j
			WHERE project_id = ? AND name = ? AND recorded_at >= ? AND recorded_at <= ?
			ORDER BY j.key ASC`,
			projectId.String(), name, from.UTC().Format(time.RFC3339Nano), to.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return nil, err
		}

		var tagKeys []string
		for tagRows.Next() {
			var key string
			if err := tagRows.Scan(&key); err != nil {
				tagRows.Close()
				return nil, err
			}
			tagKeys = append(tagKeys, key)
		}
		tagRows.Close()

		metrics = append(metrics, models.DiscoveredMetric{
			Name:    name,
			TagKeys: tagKeys,
		})
	}

	return metrics, nil
}

func (r *metricPointRepository) DiscoverTagValues(ctx context.Context, projectId uuid.UUID, metricName, tagKey string, from, to time.Time) ([]string, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT DISTINCT json_extract(tags, '$.' || ?) AS tag_value
		FROM metric_points
		WHERE project_id = ? AND name = ? AND recorded_at >= ? AND recorded_at <= ?
		AND json_extract(tags, '$.' || ?) IS NOT NULL
		AND json_extract(tags, '$.' || ?) != ''
		ORDER BY tag_value ASC`,
		tagKey, projectId.String(), metricName, from.UTC().Format(time.RFC3339Nano), to.UTC().Format(time.RFC3339Nano), tagKey, tagKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

func (r *metricPointRepository) GetAverageBetween(ctx context.Context, projectId uuid.UUID, name string, start, end time.Time) (float64, error) {
	var avg float64
	err := db.DB.QueryRowContext(ctx,
		"SELECT COALESCE(avg(value), 0) FROM metric_points WHERE project_id = ? AND name = ? AND recorded_at >= ? AND recorded_at <= ?",
		projectId.String(), name, start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano)).Scan(&avg)
	return avg, err
}

func (r *metricPointRepository) GetDistinctServers(ctx context.Context, projectId uuid.UUID, start, end time.Time) ([]string, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT DISTINCT json_extract(tags, '$.server_name') AS sn
		FROM metric_points
		WHERE project_id = ? AND recorded_at >= ? AND recorded_at <= ?
		AND json_extract(tags, '$.server_name') IS NOT NULL
		AND json_extract(tags, '$.server_name') != ''
		ORDER BY sn ASC`,
		projectId.String(), start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	return servers, nil
}

func (r *metricPointRepository) GetAverageByIntervalPerServer(ctx context.Context, projectId uuid.UUID, name string, start, end time.Time, intervalMinutes int, servers []string) (map[string][]models.TimeSeriesPoint, error) {
	secs := intervalMinutes * 60

	query := fmt.Sprintf(`SELECT
		datetime((strftime('%%s', recorded_at) / %d) * %d, 'unixepoch') AS bucket,
		json_extract(tags, '$.server_name') AS sn,
		avg(value) AS avg_value
	FROM metric_points
	WHERE project_id = ? AND name = ? AND recorded_at >= ? AND recorded_at <= ?`, secs, secs)

	args := []interface{}{projectId.String(), name, start.UTC().Format(time.RFC3339Nano), end.UTC().Format(time.RFC3339Nano)}

	if len(servers) > 0 {
		placeholders := make([]string, len(servers))
		for i, s := range servers {
			placeholders[i] = "?"
			args = append(args, s)
		}
		query += " AND json_extract(tags, '$.server_name') IN (" + strings.Join(placeholders, ", ") + ")"
	} else {
		query += " AND json_extract(tags, '$.server_name') IS NOT NULL AND json_extract(tags, '$.server_name') != ''"
	}

	query += " GROUP BY bucket, sn ORDER BY bucket ASC, sn ASC"

	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]models.TimeSeriesPoint)
	for rows.Next() {
		var bucketStr string
		var serverName string
		var value float64
		if err := rows.Scan(&bucketStr, &serverName, &value); err != nil {
			return nil, err
		}

		bucket, err := time.Parse("2006-01-02 15:04:05", bucketStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bucket time %q: %w", bucketStr, err)
		}

		result[serverName] = append(result[serverName], models.TimeSeriesPoint{
			Timestamp: bucket,
			Value:     value,
		})
	}
	return result, nil
}

func sqliteAggregationFunc(agg string) string {
	switch agg {
	case "min":
		return "min(value)"
	case "max":
		return "max(value)"
	case "sum":
		return "sum(value)"
	case "count":
		return "CAST(COUNT(*) AS REAL)"
	default:
		return "avg(value)"
	}
}

var MetricPointRepository = metricPointRepository{}
