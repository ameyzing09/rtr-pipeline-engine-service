package services

import (
	"context"
	"errors"

	"github.com/ameyzing09/rtr-pipeline-engine-service/domain"
	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/ameyzing09/rtr-pipeline-engine-service/repositories"
	"github.com/ameyzing09/rtr-pipeline-engine-service/templates"
	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/google/uuid"
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

// AssignPipeline assigns a pipeline to a job (idempotent)
// Returns nil if assignment is created or already exists with same pipeline
// Returns error if assignment exists with different pipeline (conflict)
func (s *pipelineService) AssignPipeline(ctx context.Context, pipelineID, jobID, tenantID, assignedBy string) error {
	// First validate that the pipeline belongs to this tenant
	pipeline, err := s.repo.FindByID(ctx, tenantID, pipelineID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrPipelineNotFound
		}
		return err
	}

	if pipeline.TenantID != tenantID {
		return domain.ErrPipelineNotFound // Don't leak info about other tenants' pipelines
	}

	// Check if assignment already exists
	existing, err := s.repo.FindAssignmentByJob(ctx, tenantID, jobID)
	if err == nil && existing != nil {
		// Assignment exists
		if existing.PipelineID == pipelineID {
			// Same pipeline - idempotent success
			utils.Debug("[AssignPipeline] Idempotent: job %s already assigned to pipeline %s", jobID, pipelineID)
			return nil
		}
		// Different pipeline - conflict
		utils.Warn("[AssignPipeline] Conflict: job %s already assigned to pipeline %s, cannot reassign to %s",
			jobID, existing.PipelineID, pipelineID)
		return domain.ErrDuplicatePipeline
	}

	// No existing assignment or not found - create new one
	assignment := &models.PipelineAssignment{
		PipelineID: pipelineID,
		JobID:      jobID,
		TenantID:   tenantID,
		AssignedBy: assignedBy,
	}

	return s.repo.CreateAssignment(ctx, assignment)
}

// GetAssignmentByJob retrieves the pipeline assignment for a specific job
func (s *pipelineService) GetAssignmentByJob(ctx context.Context, tenantID, jobID string) (*models.PipelineAssignment, error) {
	assignment, err := s.repo.FindAssignmentByJob(ctx, tenantID, jobID)
	if err != nil {
		// Map GORM not found error to domain error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPipelineNotFound
		}
		return nil, err
	}
	return assignment, nil
}

// SeedDefaultPipelines seeds default pipeline templates for a tenant
// It's idempotent - won't create duplicates if pipelines already exist
func (s *pipelineService) SeedDefaultPipelines(ctx context.Context, tenantID, createdBy string) (*SeedResult, error) {
	result := &SeedResult{
		Created: []string{},
		Skipped: []string{},
	}

	// Get all default templates
	templateList := templates.GetAllTemplates()

	// Track which pipelines to create
	pipelinesToCreate := []*models.Pipeline{}

	for _, template := range templateList {
		// Check if pipeline with this name already exists for this tenant
		existing, err := s.repo.FindByTenantAndName(ctx, tenantID, template.Name)

		if err == nil && existing != nil {
			// Pipeline already exists, skip it
			result.Skipped = append(result.Skipped, template.Name)
			utils.Debug("[SeedDefaults] Pipeline already exists: tenant=%s, name=%s", tenantID, template.Name)
			continue
		}

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			// Real error occurred
			utils.Error("[SeedDefaults] Error checking pipeline existence: %v", err)
			return nil, err
		}

		// Pipeline doesn't exist, prepare to create it
		pipeline, err := template.ToPipeline(tenantID, createdBy)
		if err != nil {
			utils.Error("[SeedDefaults] Error converting template to pipeline: %v", err)
			return nil, err
		}

		// Generate UUID for the pipeline
		pipeline.ID = uuid.New().String()

		pipelinesToCreate = append(pipelinesToCreate, pipeline)
		result.Created = append(result.Created, pipeline.ID)

		utils.Debug("[SeedDefaults] Prepared pipeline for creation: tenant=%s, name=%s, id=%s",
			tenantID, template.Name, pipeline.ID)
	}

	// Create all pipelines in bulk (if any)
	if len(pipelinesToCreate) > 0 {
		err := s.repo.BulkCreatePipelines(ctx, pipelinesToCreate)
		if err != nil {
			utils.Error("[SeedDefaults] Error bulk creating pipelines: %v", err)
			return nil, err
		}
		utils.Info("[SeedDefaults] Successfully created %d default pipelines for tenant %s", len(pipelinesToCreate), tenantID)
	} else {
		utils.Info("[SeedDefaults] All default pipelines already exist for tenant %s", tenantID)
	}

	return result, nil
}
