package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/domain"
	"github.com/ameyzing09/rtr-pipeline-engine-service/middleware"
	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/ameyzing09/rtr-pipeline-engine-service/services"
	"github.com/ameyzing09/rtr-pipeline-engine-service/utils/httpx"
	"github.com/gin-gonic/gin"
)

type PipelineHandler struct {
	pipelineService services.PipelineService
}

func NewPipelineHandler(pipelineService services.PipelineService) *PipelineHandler {
	return &PipelineHandler{
		pipelineService: pipelineService,
	}
}

func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	var body CreatePipelineDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	// Get request context (contains tenant ID and user ID from JWT)
	requestCtx, exists := middleware.GetRequestContext(c)
	if !exists {
		httpx.HandleError(c, domain.ErrMissingTenantID)
		return
	}

	stagesJSON, err := json.Marshal(body.Stages)
	if err != nil {
		httpx.HandleError(c, domain.ErrValidationFailed)
		return
	}

	pipeline := models.Pipeline{
		TenantID:    requestCtx.TenantID,
		Name:        body.Name,
		Description: body.Description,
		Stages:      stagesJSON,
		CreatedBy:   requestCtx.UserID,
	}

	if err := h.pipelineService.CreatePipeline(c.Request.Context(), &pipeline); err != nil {
		httpx.HandleError(c, domain.ErrDatabaseOperation)
		return
	}

	c.JSON(http.StatusOK, pipeline)
}

func (h *PipelineHandler) ListPipelines(c *gin.Context) {
	// Get request context (contains tenant ID from JWT)
	requestCtx, exists := middleware.GetRequestContext(c)
	if !exists {
		httpx.HandleError(c, domain.ErrMissingTenantID)
		return
	}

	pipelines, err := h.pipelineService.ListPipelines(c.Request.Context(), requestCtx.TenantID)
	if err != nil {
		httpx.HandleError(c, domain.ErrDatabaseOperation)
		return
	}

	c.JSON(http.StatusOK, pipelines)
}

func (h *PipelineHandler) AssignPipeline(c *gin.Context) {
	var pipelineBody PipelineAssignmentDTO
	if err := c.ShouldBindJSON(&pipelineBody); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	// Get request context (contains tenant ID and user ID from JWT)
	requestCtx, exists := middleware.GetRequestContext(c)
	if !exists {
		httpx.HandleError(c, domain.ErrMissingTenantID)
		return
	}

	if err := h.pipelineService.AssignPipeline(c.Request.Context(), pipelineBody.PipelineID, pipelineBody.JobID, requestCtx.TenantID, requestCtx.UserID); err != nil {
		httpx.HandleError(c, domain.ErrDatabaseOperation)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pipeline assigned successfully"})
}
