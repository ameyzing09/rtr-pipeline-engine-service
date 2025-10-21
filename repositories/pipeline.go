package repositories

import (
	"context"

	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"gorm.io/gorm"
)

// PipelineRepository defines the interface for pipeline data access
type PipelineRepository interface {
	Create(ctx context.Context, pipeline *models.Pipeline) error
	FindByTenant(ctx context.Context, tenantID string) ([]models.Pipeline, error)
	FindByID(ctx context.Context, tenantID, pipelineID string) (*models.Pipeline, error)
	CreateAssignment(ctx context.Context, assignment *models.PipelineAssignment) error
}

// gormPipelineRepo is the concrete GORM implementation of PipelineRepository
type gormPipelineRepo struct {
	db *gorm.DB
}

// NewPipelineRepository creates a new pipeline repository
func NewPipelineRepository(db *gorm.DB) PipelineRepository {
	return &gormPipelineRepo{db: db}
}

// Create creates a new pipeline
func (r *gormPipelineRepo) Create(ctx context.Context, pipeline *models.Pipeline) error {
	// Check for duplicate pipeline name within tenant
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Pipeline{}).
		Where("tenant_id = ? AND name = ? AND is_deleted = ?", pipeline.TenantID, pipeline.Name, false).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	return r.db.WithContext(ctx).Create(pipeline).Error
}

// FindByTenant finds all active pipelines for a tenant
func (r *gormPipelineRepo) FindByTenant(ctx context.Context, tenantID string) ([]models.Pipeline, error) {
	var pipelines []models.Pipeline
	err := r.db.WithContext(ctx).
		Where(map[string]interface{}{
			"tenant_id":  tenantID,
			"is_deleted": false,
		}).
		Find(&pipelines).Error

	return pipelines, err
}

// FindByID finds a pipeline by ID and tenant
func (r *gormPipelineRepo) FindByID(ctx context.Context, tenantID, pipelineID string) (*models.Pipeline, error) {
	var pipeline models.Pipeline
	err := r.db.WithContext(ctx).
		Where(map[string]interface{}{
			"id":         pipelineID,
			"tenant_id":  tenantID,
			"is_deleted": false,
		}).
		First(&pipeline).Error

	if err != nil {
		return nil, err
	}

	return &pipeline, nil
}

// CreateAssignment creates a new pipeline assignment
func (r *gormPipelineRepo) CreateAssignment(ctx context.Context, assignment *models.PipelineAssignment) error {
	// Check for duplicate assignment (one pipeline per job per tenant)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.PipelineAssignment{}).
		Where("tenant_id = ? AND job_id = ? AND is_deleted = ?", assignment.TenantID, assignment.JobID, false).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return gorm.ErrDuplicatedKey
	}

	return r.db.WithContext(ctx).Create(assignment).Error
}
