package services

import (
	"context"

	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/ameyzing09/rtr-pipeline-engine-service/repositories"
)

type pipelineService struct {
	repo repositories.PipelineRepository
}

// NewPipelineService creates a new pipeline service
func NewPipelineService(repo repositories.PipelineRepository) PipelineService {
	return &pipelineService{repo: repo}
}

// CreatePipeline creates a new pipeline
func (s *pipelineService) CreatePipeline(ctx context.Context, pipeline *models.Pipeline) error {
	// No business rule validation - tenants have full flexibility
	// Only basic JSON validation happens at handler layer
	return s.repo.Create(ctx, pipeline)
}

// ListPipelines returns all pipelines for a tenant
func (s *pipelineService) ListPipelines(ctx context.Context, tenantID string) ([]models.Pipeline, error) {
	return s.repo.FindByTenant(ctx, tenantID)
}

// AssignPipeline assigns a pipeline to a job
func (s *pipelineService) AssignPipeline(ctx context.Context, pipelineID, jobID, tenantID, assignedBy string) error {
	assignment := &models.PipelineAssignment{
		PipelineID: pipelineID,
		JobID:      jobID,
		TenantID:   tenantID,
		AssignedBy: assignedBy,
	}

	return s.repo.CreateAssignment(ctx, assignment)
}
