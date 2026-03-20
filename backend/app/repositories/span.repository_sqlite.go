//go:build !pgch

package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
)

type spanRepository struct{}

func (r *spanRepository) InsertAsync(ctx context.Context, spans []models.Span) error {
	if len(spans) == 0 {
		return nil
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO spans (id, trace_id, project_id, name, start_time, duration, recorded_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range spans {
		if _, err := stmt.ExecContext(ctx,
			s.Id.String(), s.TraceId.String(), s.ProjectId.String(),
			s.Name, s.StartTime.UTC().Format(time.RFC3339Nano),
			int64(s.Duration), s.RecordedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *spanRepository) FindByTraceId(ctx context.Context, projectId, traceId uuid.UUID) ([]models.Span, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT id, trace_id, project_id, name, start_time, duration, recorded_at
		FROM spans
		WHERE project_id = ? AND trace_id = ?
		ORDER BY start_time ASC`,
		projectId.String(), traceId.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spans []models.Span
	for rows.Next() {
		var s models.Span
		var idStr, traceIdStr, projectIdStr string
		var startTimeStr, recordedAtStr string
		if err := rows.Scan(&idStr, &traceIdStr, &projectIdStr, &s.Name, &startTimeStr, &s.Duration, &recordedAtStr); err != nil {
			return nil, err
		}
		s.Id, _ = uuid.Parse(idStr)
		s.TraceId, _ = uuid.Parse(traceIdStr)
		s.ProjectId, _ = uuid.Parse(projectIdStr)
		s.StartTime, _ = time.Parse(time.RFC3339Nano, startTimeStr)
		s.RecordedAt, _ = time.Parse(time.RFC3339Nano, recordedAtStr)
		spans = append(spans, s)
	}

	return spans, nil
}

var SpanRepository = spanRepository{}
