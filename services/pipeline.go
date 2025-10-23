package services

import (
	"context"
	"errors"

	"github.com/ameyzing09/rtr-pipeline-engine-service/domain"
	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/ameyzing09/rtr-pipeline-engine-service/repositories"
	"gorm.io/gorm"
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
	err := s.repo.Create(ctx, pipeline)
	if err != nil {
		// Map duplicate key error to domain error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrDuplicatePipeline
		}
		return err
	}
	return nil
}

// GetPipelineByID returns a single pipeline by ID for a tenant
func (s *pipelineService) GetPipelineByID(ctx context.Context, tenantID, pipelineID string) (*models.Pipeline, error) {
	pipeline, err := s.repo.FindByID(ctx, tenantID, pipelineID)
	if err != nil {
		// Map GORM not found error to domain error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPipelineNotFound
		}
		return nil, err
	}
	return pipeline, nil
}

// UpdatePipeline updates an existing pipeline
func (s *pipelineService) UpdatePipeline(ctx context.Context, pipeline *models.Pipeline) error {
	err := s.repo.Update(ctx, pipeline)
	if err != nil {
		// Map GORM errors to domain errors
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrPipelineNotFound
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrDuplicatePipeline
		}
		return err
	}
	return nil
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
