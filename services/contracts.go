package services

import (
	"context"

	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
)

// SeedResult contains the result of seeding default pipelines
type SeedResult struct {
	Created []string // IDs of created pipelines
	Skipped []string // Names of skipped pipelines (already exist)
}

// PipelineService defines the interface for pipeline business logic
type PipelineService interface {
	CreatePipeline(ctx context.Context, pipeline *models.Pipeline) error
	GetPipelineByID(ctx context.Context, tenantID, pipelineID string) (*models.Pipeline, error)
	UpdatePipeline(ctx context.Context, pipeline *models.Pipeline) error
	ListPipelines(ctx context.Context, tenantID string) ([]models.Pipeline, error)
	AssignPipeline(ctx context.Context, pipelineID, jobID, tenantID, assignedBy string) error
	GetAssignmentByJob(ctx context.Context, tenantID, jobID string) (*models.PipelineAssignment, error)
	SeedDefaultPipelines(ctx context.Context, tenantID, createdBy string) (*SeedResult, error)
}
