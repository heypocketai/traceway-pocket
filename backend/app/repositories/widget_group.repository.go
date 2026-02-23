package repositories

import (
	"database/sql"

	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"

	"github.com/google/uuid"
	"github.com/tracewayapp/lit/v2"
)

type widgetGroupRepository struct{}

func (r *widgetGroupRepository) FindByProject(tx *sql.Tx, projectId uuid.UUID) ([]*models.WidgetGroup, error) {
	return lit.SelectNamed[models.WidgetGroup](
		tx,
		"SELECT id, project_id, name, description, is_default, created_by, created_at, updated_at FROM widget_groups WHERE project_id = :project_id ORDER BY is_default DESC, name ASC",
		lit.P{"project_id": projectId},
	)
}

func (r *widgetGroupRepository) FindById(tx *sql.Tx, id int) (*models.WidgetGroup, error) {
	return lit.SelectSingleNamed[models.WidgetGroup](
		tx,
		"SELECT id, project_id, name, description, is_default, created_by, created_at, updated_at FROM widget_groups WHERE id = :id",
		lit.P{"id": id},
	)
}

func (r *widgetGroupRepository) Create(tx *sql.Tx, group *models.WidgetGroup) (int, error) {
	return lit.Insert[models.WidgetGroup](tx, group)
}

func (r *widgetGroupRepository) Update(tx *sql.Tx, group *models.WidgetGroup) error {
	return lit.UpdateNamed(tx, group, "id = :id", lit.P{"id": group.Id})
}

func (r *widgetGroupRepository) Delete(tx *sql.Tx, id int) error {
	return lit.DeleteNamed(db.Driver, tx, "DELETE FROM widget_groups WHERE id = :id", lit.P{"id": id})
}

func (r *widgetGroupRepository) FindWidgetsByGroup(tx *sql.Tx, widgetGroupId int) ([]*models.WidgetGroupWidget, error) {
	return lit.SelectNamed[models.WidgetGroupWidget](
		tx,
		"SELECT id, widget_group_id, title, widget_type, config, position, created_at, updated_at FROM widget_group_widgets WHERE widget_group_id = :wg_id ORDER BY position ASC",
		lit.P{"wg_id": widgetGroupId},
	)
}

func (r *widgetGroupRepository) FindWidgetById(tx *sql.Tx, id int) (*models.WidgetGroupWidget, error) {
	return lit.SelectSingleNamed[models.WidgetGroupWidget](
		tx,
		"SELECT id, widget_group_id, title, widget_type, config, position, created_at, updated_at FROM widget_group_widgets WHERE id = :id",
		lit.P{"id": id},
	)
}

func (r *widgetGroupRepository) CreateWidget(tx *sql.Tx, widget *models.WidgetGroupWidget) (int, error) {
	return lit.Insert[models.WidgetGroupWidget](tx, widget)
}

func (r *widgetGroupRepository) UpdateWidget(tx *sql.Tx, widget *models.WidgetGroupWidget) error {
	return lit.UpdateNamed(tx, widget, "id = :id", lit.P{"id": widget.Id})
}

func (r *widgetGroupRepository) DeleteWidget(tx *sql.Tx, id int) error {
	return lit.DeleteNamed(db.Driver, tx, "DELETE FROM widget_group_widgets WHERE id = :id", lit.P{"id": id})
}

var WidgetGroupRepository = widgetGroupRepository{}
