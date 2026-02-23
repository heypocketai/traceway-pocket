package models

import (
	"github.com/tracewayapp/traceway/backend/app/config"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	Id             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Token          string    `json:"token"`
	Framework      string    `json:"framework"`
	OrganizationId *int      `json:"organizationId"`
	CreatedAt      time.Time `json:"createdAt"`
	SourceMapToken *string   `json:"sourceMapToken,omitempty"`
}

func (p Project) ToProjectWithBackendUrl() *ProjectWithBackendUrl {
	return &ProjectWithBackendUrl{Project: p, BackendUrl: getBackendUrl()}
}

func getBackendUrl() string {
	if url := config.Config.AppBaseURL; url != "" {
		return url
	}
	return "https://cloud.tracewayapp.com"
}

type ProjectWithBackendUrl struct {
	Project
	BackendUrl string `json:"backendUrl"`
}
