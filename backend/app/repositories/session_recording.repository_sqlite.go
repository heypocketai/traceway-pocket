//go:build !pgch

package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"
)

type sessionRecordingRepository struct{}

func (r *sessionRecordingRepository) InsertAsync(ctx context.Context, recordings []models.SessionRecording) error {
	if len(recordings) == 0 {
		return nil
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO session_recordings (id, project_id, exception_id, file_path, recorded_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, rec := range recordings {
		if _, err := stmt.ExecContext(ctx,
			rec.Id.String(), rec.ProjectId.String(), rec.ExceptionId.String(),
			rec.FilePath, rec.RecordedAt.UTC().Format(time.RFC3339Nano),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sessionRecordingRepository) FindByExceptionId(ctx context.Context, projectId uuid.UUID, exceptionId uuid.UUID) (string, error) {
	var filePath string
	err := db.DB.QueryRowContext(ctx,
		"SELECT file_path FROM session_recordings WHERE project_id = ? AND exception_id = ? ORDER BY recorded_at DESC LIMIT 1",
		projectId.String(), exceptionId.String()).Scan(&filePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return filePath, nil
}

var SessionRecordingRepository = sessionRecordingRepository{}
