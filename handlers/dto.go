package handlers

// Stage represents a pipeline stage
// Fully customizable by tenants - no restrictions on values
type Stage struct {
	// Stage name (required, 1-100 characters)
	// Example: "Phone Screen", "Technical Round", "Culture Fit"
	Stage string `json:"stage" binding:"required,min=1,max=100"`

	// Stage type (required, 1-50 characters)
	// Can be predefined (phone, technical, system_design, behavioral, offer)
	// OR custom tenant-defined types (e.g., "case_study", "portfolio_review")
	Type string `json:"type" binding:"required,min=1,max=50"`

	// Conducted by (required, 1-50 characters)
	// Can be predefined (auto, hr, interviewer)
	// OR custom tenant-defined roles (e.g., "hiring_manager", "team_lead", "cto")
	ConductedBy string `json:"conducted_by" binding:"required,min=1,max=50"`

	// Custom metadata (optional)
	// Tenants can add any additional fields like:
	// - duration_minutes, scoring_criteria, required_skills, etc.
	Metadata map[string]interface{} `json:"metadata" binding:"omitempty"`
}

// CreatePipelineDTO for creating a new pipeline
type CreatePipelineDTO struct {
	// Pipeline name (required, 3-255 characters)
	Name string `json:"name" binding:"required,min=3,max=255"`

	// Pipeline description (optional, max 1000 characters)
	Description string `json:"description" binding:"omitempty,max=1000"`

	// Pipeline stages (required, at least 1 stage, max 10 stages)
	// Each stage must be valid (nested validation with 'dive')
	Stages []Stage `json:"stages" binding:"required,min=1,max=10,dive"`

	// Custom tenant-specific fields (optional, will be validated against tenant schema)
	Extra map[string]interface{} `json:"extra" binding:"omitempty"`
}

// PipelineAssignmentDTO for assigning a pipeline to a job
type PipelineAssignmentDTO struct {
	// Pipeline ID (required, must be valid UUID format)
	PipelineID string `json:"pipeline_id" binding:"required,uuid"`

	// Job ID (required, must be valid UUID format)
	JobID string `json:"job_id" binding:"required,uuid"`
}
