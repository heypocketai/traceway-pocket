package repositories

import (
	"backend/app/models"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tracewayapp/lit"
)

type metricRegistryRepository struct {
	knownMetrics sync.Map
}

func (r *metricRegistryRepository) EnsureRegistered(tx *sql.Tx, projectId uuid.UUID, names []string) error {
	for _, name := range names {
		key := projectId.String() + ":" + name
		if _, loaded := r.knownMetrics.Load(key); loaded {
			continue
		}

		metricType := defaultMetricType(name)
		unit := defaultUnit(name)

		entry := &models.MetricRegistry{
			ProjectId:   projectId,
			Name:        name,
			MetricType:  metricType,
			Unit:        unit,
			Description: "",
			CreatedAt:   time.Now().UTC(),
		}

		_, err := lit.Insert[models.MetricRegistry](tx, entry)
		if err != nil {
			existing, findErr := r.FindByProjectAndName(tx, projectId, name)
			if findErr != nil || existing == nil {
				return err
			}
		}
		r.knownMetrics.Store(key, true)
	}
	return nil
}

func (r *metricRegistryRepository) FindByProject(tx *sql.Tx, projectId uuid.UUID) ([]*models.MetricRegistry, error) {
	return lit.Select[models.MetricRegistry](
		tx,
		"SELECT id, project_id, name, metric_type, unit, description, created_at FROM metric_registry WHERE project_id = $1 ORDER BY name ASC",
		projectId,
	)
}

func (r *metricRegistryRepository) FindByProjectAndName(tx *sql.Tx, projectId uuid.UUID, name string) (*models.MetricRegistry, error) {
	return lit.SelectSingle[models.MetricRegistry](
		tx,
		"SELECT id, project_id, name, metric_type, unit, description, created_at FROM metric_registry WHERE project_id = $1 AND name = $2",
		projectId, name,
	)
}

func (r *metricRegistryRepository) Update(tx *sql.Tx, entry *models.MetricRegistry) error {
	return lit.Update(tx, entry, "id = $1", entry.Id)
}

func defaultMetricType(name string) string {
	switch name {
	case models.MetricNameNumGC:
		return "counter"
	case models.MetricNameGCPauseTotal:
		return "counter"
	default:
		return "gauge"
	}
}

func defaultUnit(name string) string {
	switch name {
	case models.MetricNameCpuUsage:
		return "%"
	case models.MetricNameMemoryUsage, models.MetricNameMemoryTotal:
		return "MB"
	case models.MetricNameGoRoutines, models.MetricNameHeapObjects:
		return "count"
	case models.MetricNameNumGC:
		return "count"
	case models.MetricNameGCPauseTotal:
		return "ns"
	default:
		return ""
	}
}

var MetricRegistryRepository = metricRegistryRepository{}
