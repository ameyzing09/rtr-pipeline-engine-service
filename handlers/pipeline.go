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

func (h *PipelineHandler) GetPipelineByID(c *gin.Context) {
	// Get request context (contains tenant ID from JWT)
	requestCtx, exists := middleware.GetRequestContext(c)
	if !exists {
		httpx.HandleError(c, domain.ErrMissingTenantID)
		return
	}

	// Get pipeline ID from URL parameter
	pipelineID := c.Param("id")
	if pipelineID == "" {
		httpx.HandleError(c, domain.ErrInvalidRequest)
		return
	}

	pipeline, err := h.pipelineService.GetPipelineByID(c.Request.Context(), requestCtx.TenantID, pipelineID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	httpx.RespondWithSuccess(c, http.StatusOK, pipeline)
}

func (h *PipelineHandler) UpdatePipeline(c *gin.Context) {
	// Get request context (contains tenant ID and user ID from JWT)
	requestCtx, exists := middleware.GetRequestContext(c)
	if !exists {
		httpx.HandleError(c, domain.ErrMissingTenantID)
		return
	}

	// Get pipeline ID from URL parameter
	pipelineID := c.Param("id")
	if pipelineID == "" {
		httpx.HandleError(c, domain.ErrInvalidRequest)
		return
	}

	// Bind and validate update DTO
	var updateBody UpdatePipelineDTO
	if err := c.ShouldBindJSON(&updateBody); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	// Fetch existing pipeline to merge updates
	existing, err := h.pipelineService.GetPipelineByID(c.Request.Context(), requestCtx.TenantID, pipelineID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Merge updates: only update fields that are provided
	if updateBody.Name != nil {
		existing.Name = *updateBody.Name
	}
	if updateBody.Description != nil {
		existing.Description = *updateBody.Description
	}
	if updateBody.Stages != nil {
		stagesJSON, err := json.Marshal(*updateBody.Stages)
		if err != nil {
			httpx.HandleError(c, domain.ErrValidationFailed)
			return
		}
		existing.Stages = stagesJSON
	}

	// Set updated_by from JWT context
	existing.UpdatedBy = requestCtx.UserID

	// Update the pipeline
	if err := h.pipelineService.UpdatePipeline(c.Request.Context(), existing); err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Return updated pipeline
	httpx.RespondWithSuccess(c, http.StatusOK, existing)
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
