package pipeline_dto

type Stage struct {
	Stage string `json:"stage" binding:"required"`
	Type  string `json:"type" binding:"required"`
}

type CreatePipelineDTO struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Stages      []Stage `json:"stages" binding:"required"`
}

type PipelineAssignmentDTO struct {
	PipelineID string `json:"pipeline_id" binding:"required"`
	JobID      string `json:"job_id" binding:"required"`
}
