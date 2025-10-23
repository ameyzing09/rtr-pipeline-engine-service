package services

import (
	"context"

	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
)

// PipelineService defines the interface for pipeline business logic
type PipelineService interface {
	CreatePipeline(ctx context.Context, pipeline *models.Pipeline) error
	GetPipelineByID(ctx context.Context, tenantID, pipelineID string) (*models.Pipeline, error)
	UpdatePipeline(ctx context.Context, pipeline *models.Pipeline) error
	ListPipelines(ctx context.Context, tenantID string) ([]models.Pipeline, error)
	AssignPipeline(ctx context.Context, pipelineID, jobID, tenantID, assignedBy string) error
}
