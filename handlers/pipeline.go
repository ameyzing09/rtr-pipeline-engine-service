package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/domain"
	"github.com/ameyzing09/rtr-pipeline-engine-service/middleware"
	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/ameyzing09/rtr-pipeline-engine-service/services"
	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
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

// GetPipelineAssignment retrieves the pipeline assigned to a specific job
func (h *PipelineHandler) GetPipelineAssignment(c *gin.Context) {
	// Get request context (contains tenant ID from JWT)
	requestCtx, exists := middleware.GetRequestContext(c)
	if !exists {
		httpx.HandleError(c, domain.ErrMissingTenantID)
		return
	}

	// Get job_id from query parameter
	jobID := c.Query("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job_id query parameter is required",
			"code":  "MISSING_JOB_ID",
		})
		return
	}

	// Get assignment from service
	assignment, err := h.pipelineService.GetAssignmentByJob(c.Request.Context(), requestCtx.TenantID, jobID)
	if err != nil {
		if errors.Is(err, domain.ErrPipelineNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No pipeline assignment found for this job",
				"code":  "ASSIGNMENT_NOT_FOUND",
			})
			return
		}
		httpx.HandleError(c, err)
		return
	}

	// Log the query for audit
	utils.Debug("[GetAssignment] tenant=%s, user=%s, job=%s, pipeline=%s",
		requestCtx.TenantID, requestCtx.UserID, jobID, assignment.PipelineID)

	// Return pipeline_id
	c.JSON(http.StatusOK, gin.H{
		"pipeline_id": assignment.PipelineID,
		"job_id":      assignment.JobID,
		"assigned_at": assignment.CreatedAt,
	})
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

	// Check if assignment already exists (for idempotency detection)
	existingAssignment, _ := h.pipelineService.GetAssignmentByJob(c.Request.Context(), requestCtx.TenantID, pipelineBody.JobID)
	alreadyExists := existingAssignment != nil

	// Attempt to assign pipeline
	err := h.pipelineService.AssignPipeline(c.Request.Context(), pipelineBody.PipelineID, pipelineBody.JobID, requestCtx.TenantID, requestCtx.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicatePipeline) {
			// Conflict: job already assigned to different pipeline
			c.JSON(http.StatusConflict, gin.H{
				"error": "Job is already assigned to a different pipeline",
				"code":  "PIPELINE_ASSIGNMENT_CONFLICT",
			})
			return
		}
		if errors.Is(err, domain.ErrPipelineNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Pipeline not found or does not belong to this tenant",
				"code":  "PIPELINE_NOT_FOUND",
			})
			return
		}
		httpx.HandleError(c, domain.ErrDatabaseOperation)
		return
	}

	// Determine status code based on whether it was created or already existed
	statusCode := http.StatusCreated
	message := "Pipeline assigned successfully"
	if alreadyExists {
		statusCode = http.StatusOK
		message = "Pipeline assignment already exists (idempotent)"
	}

	utils.Info("[AssignPipeline] tenant=%s, user=%s, job=%s, pipeline=%s, status=%d",
		requestCtx.TenantID, requestCtx.UserID, pipelineBody.JobID, pipelineBody.PipelineID, statusCode)

	c.JSON(statusCode, gin.H{
		"message":     message,
		"pipeline_id": pipelineBody.PipelineID,
		"job_id":      pipelineBody.JobID,
	})
}

// SeedDefaults seeds default pipeline templates for a tenant
// Supports two modes:
//   - Self-seed: ADMIN/HR can seed their own tenant (X-Tenant-Id == JWT.tid)
//   - Cross-tenant: SUPERADMIN can seed any tenant via X-Tenant-Id
//
// Idempotency: Supports X-Idempotency-Key header for safe retries
func (h *PipelineHandler) SeedDefaults(c *gin.Context) {
	// Get request context (contains tenant ID and user ID from JWT)
	requestCtx, exists := middleware.GetRequestContext(c)
	if !exists {
		httpx.HandleError(c, domain.ErrMissingTenantID)
		return
	}

	// Extract idempotency key (optional)
	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	if idempotencyKey != "" {
		utils.Debug("[SeedDefaults] Idempotency key provided: %s", idempotencyKey)
		// TODO: Implement full idempotency key support with cache/store
		// For now, the unique constraint on (tenant_id, name) provides basic idempotency
	}

	// Get the effective tenant ID from X-Tenant-Id header (set by CrossTenantMiddleware)
	// This will be either:
	// - The user's own tenant (for ADMIN/HR self-seed, validated by TenantMiddleware)
	// - Any tenant (for SUPERADMIN cross-tenant, SUPERADMIN bypass in TenantMiddleware)
	targetTenantID, exists := middleware.GetTargetTenantID(c)
	if !exists || targetTenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Target tenant ID is required",
			"code":  "MISSING_TENANT_ID",
		})
		return
	}

	// Call service to seed defaults
	result, err := h.pipelineService.SeedDefaultPipelines(c.Request.Context(), targetTenantID, requestCtx.UserID)
	if err != nil {
		httpx.HandleError(c, domain.ErrDatabaseOperation)
		return
	}

	// Determine response status code
	statusCode := http.StatusOK
	if len(result.Created) > 0 {
		statusCode = http.StatusCreated
	}

	// Build response message
	message := "All default pipelines already exist"
	if len(result.Created) > 0 {
		message = "Successfully created default pipelines"
	}

	// Comprehensive audit logging
	// TODO: Consider persisting these fields in pipelines table: created_by_role, created_by_tenant_id
	utils.Info("[SeedDefaults] Audit: request_id=%s, requested_by_uid=%s, requested_by_role=%s, "+
		"requested_by_tenant=%s, effective_tenant=%s, created=%d, skipped=%d, status=%d",
		c.GetString("request_id"),
		requestCtx.UserID,
		requestCtx.Role,
		requestCtx.TenantID,
		targetTenantID,
		len(result.Created),
		len(result.Skipped),
		statusCode)

	c.JSON(statusCode, gin.H{
		"created":    result.Created,
		"skipped":    result.Skipped,
		"tenant_id":  targetTenantID,
		"seeded_by":  requestCtx.UserID,
		"message":    message,
	})
}
