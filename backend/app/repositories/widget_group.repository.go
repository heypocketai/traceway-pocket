package repositories

import (
	"backend/app/models"
	"database/sql"

	"github.com/google/uuid"
	"github.com/tracewayapp/lit"
)

type widgetGroupRepository struct{}

func (r *widgetGroupRepository) FindByProject(tx *sql.Tx, projectId uuid.UUID) ([]*models.WidgetGroup, error) {
	return lit.Select[models.WidgetGroup](
		tx,
		"SELECT id, project_id, name, description, is_default, created_by, created_at, updated_at FROM widget_groups WHERE project_id = $1 ORDER BY is_default DESC, name ASC",
		projectId,
	)
}

func (r *widgetGroupRepository) FindById(tx *sql.Tx, id int) (*models.WidgetGroup, error) {
	return lit.SelectSingle[models.WidgetGroup](
		tx,
		"SELECT id, project_id, name, description, is_default, created_by, created_at, updated_at FROM widget_groups WHERE id = $1",
		id,
	)
}

func (r *widgetGroupRepository) Create(tx *sql.Tx, group *models.WidgetGroup) (int, error) {
	return lit.Insert[models.WidgetGroup](tx, group)
}

func (r *widgetGroupRepository) Update(tx *sql.Tx, group *models.WidgetGroup) error {
	return lit.Update(tx, group, "id = $1", group.Id)
}

func (r *widgetGroupRepository) Delete(tx *sql.Tx, id int) error {
	return lit.Delete(tx, "DELETE FROM widget_groups WHERE id = $1", id)
}

func (r *widgetGroupRepository) FindWidgetsByGroup(tx *sql.Tx, widgetGroupId int) ([]*models.WidgetGroupWidget, error) {
	return lit.Select[models.WidgetGroupWidget](
		tx,
		"SELECT id, widget_group_id, title, widget_type, config, position, created_at, updated_at FROM widget_group_widgets WHERE widget_group_id = $1 ORDER BY position ASC",
		widgetGroupId,
	)
}

func (r *widgetGroupRepository) FindWidgetById(tx *sql.Tx, id int) (*models.WidgetGroupWidget, error) {
	return lit.SelectSingle[models.WidgetGroupWidget](
		tx,
		"SELECT id, widget_group_id, title, widget_type, config, position, created_at, updated_at FROM widget_group_widgets WHERE id = $1",
		id,
	)
}

func (r *widgetGroupRepository) CreateWidget(tx *sql.Tx, widget *models.WidgetGroupWidget) (int, error) {
	return lit.Insert[models.WidgetGroupWidget](tx, widget)
}

func (r *widgetGroupRepository) UpdateWidget(tx *sql.Tx, widget *models.WidgetGroupWidget) error {
	return lit.Update(tx, widget, "id = $1", widget.Id)
}

func (r *widgetGroupRepository) DeleteWidget(tx *sql.Tx, id int) error {
	return lit.Delete(tx, "DELETE FROM widget_group_widgets WHERE id = $1", id)
}

var WidgetGroupRepository = widgetGroupRepository{}
