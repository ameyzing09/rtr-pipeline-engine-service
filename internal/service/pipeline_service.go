package service

import (
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/model"
	"gorm.io/gorm"
)

type PipelineService interface {
	CreatePipeline(p *model.Pipeline) error
	ListPipelines(tenantId string) ([]model.Pipeline, error)
}

type pipelineService struct {
	db *gorm.DB
}

func NewPipelineService(db *gorm.DB) PipelineService {
	return &pipelineService{db: db}
}

func (s *pipelineService) CreatePipeline(p *model.Pipeline) error {
	return s.db.Create(p).Error
}

func (s *pipelineService) ListPipelines(tenantId string) ([]model.Pipeline, error) {
	var pipelines []model.Pipeline
	err := s.db.Where(map[string]any{
		"tenant_id":  tenantId,
		"is_deleted": false,
	}).Find(&pipelines).Error

	return pipelines, err
}
