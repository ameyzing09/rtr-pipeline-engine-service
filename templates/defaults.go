package templates

import (
	"encoding/json"

	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"gorm.io/datatypes"
)

// Stage represents a pipeline stage
type Stage struct {
	Stage       string                 `json:"stage"`
	Type        string                 `json:"type"`
	ConductedBy string                 `json:"conducted_by"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PipelineTemplate represents a default pipeline template
type PipelineTemplate struct {
	Name        string
	Description string
	Stages      []Stage
}

// GetStandardPipeline returns the standard 3-stage pipeline template
// Suitable for most hiring workflows
func GetStandardPipeline() *PipelineTemplate {
	return &PipelineTemplate{
		Name:        "Standard Hiring Pipeline",
		Description: "A standard 3-stage hiring pipeline suitable for most positions",
		Stages: []Stage{
			{
				Stage:       "Application Review",
				Type:        "manual_review",
				ConductedBy: "hr",
				Metadata: map[string]interface{}{
					"duration_minutes": 15,
					"description":      "Initial review of application and resume",
				},
			},
			{
				Stage:       "Technical Interview",
				Type:        "technical",
				ConductedBy: "interviewer",
				Metadata: map[string]interface{}{
					"duration_minutes": 60,
					"description":      "Technical skills assessment and problem-solving",
				},
			},
			{
				Stage:       "Final Interview",
				Type:        "final",
				ConductedBy: "hiring_manager",
				Metadata: map[string]interface{}{
					"duration_minutes": 45,
					"description":      "Final round with hiring manager",
				},
			},
		},
	}
}

// GetComprehensivePipeline returns the comprehensive 5-stage pipeline template
// Suitable for senior positions or thorough evaluation processes
func GetComprehensivePipeline() *PipelineTemplate {
	return &PipelineTemplate{
		Name:        "Comprehensive Hiring Pipeline",
		Description: "A comprehensive 5-stage hiring pipeline for thorough candidate evaluation",
		Stages: []Stage{
			{
				Stage:       "Application Review",
				Type:        "manual_review",
				ConductedBy: "hr",
				Metadata: map[string]interface{}{
					"duration_minutes": 15,
					"description":      "Initial review of application and resume",
				},
			},
			{
				Stage:       "Phone Screening",
				Type:        "phone",
				ConductedBy: "hr",
				Metadata: map[string]interface{}{
					"duration_minutes": 30,
					"description":      "Initial phone screening to assess basic fit",
				},
			},
			{
				Stage:       "Technical Assessment",
				Type:        "technical",
				ConductedBy: "interviewer",
				Metadata: map[string]interface{}{
					"duration_minutes": 90,
					"description":      "In-depth technical skills assessment",
				},
			},
			{
				Stage:       "Behavioral Interview",
				Type:        "behavioral",
				ConductedBy: "interviewer",
				Metadata: map[string]interface{}{
					"duration_minutes": 60,
					"description":      "Cultural fit and behavioral assessment",
				},
			},
			{
				Stage:       "Final Interview",
				Type:        "final",
				ConductedBy: "hiring_manager",
				Metadata: map[string]interface{}{
					"duration_minutes": 45,
					"description":      "Final decision round with senior leadership",
				},
			},
		},
	}
}

// GetAllTemplates returns all default pipeline templates
func GetAllTemplates() []*PipelineTemplate {
	return []*PipelineTemplate{
		GetStandardPipeline(),
		GetComprehensivePipeline(),
	}
}

// ToPipeline converts a PipelineTemplate to a models.Pipeline
// tenantID and createdBy must be provided
func (pt *PipelineTemplate) ToPipeline(tenantID, createdBy string) (*models.Pipeline, error) {
	// Serialize stages to JSON
	stagesJSON, err := json.Marshal(pt.Stages)
	if err != nil {
		return nil, err
	}

	return &models.Pipeline{
		TenantID:    tenantID,
		Name:        pt.Name,
		Description: pt.Description,
		Stages:      datatypes.JSON(stagesJSON),
		IsActive:    true,
		IsDeleted:   false,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}, nil
}
