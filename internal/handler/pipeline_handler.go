package handler

import (
	"net/http"

	"encoding/json"

	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/dto"
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/model"
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/service"
	"github.com/gin-gonic/gin"
)

type PipelineHandler struct {
	PipelineService service.PipelineService
}

func NewPipelineHandler(plService service.PipelineService) *PipelineHandler {
	return &PipelineHandler{
		PipelineService: plService,
	}
}

func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	var body dto.CreatePipelineDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenantId := c.GetString("tenantId")
	stagesJson, err := json.Marshal(body.Stages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal stages"})
		return
	}

	pipeline := model.Pipeline{
		TenantID:    tenantId,
		Name:        body.Name,
		Description: body.Description,
		Stages:      stagesJson,
	}

	if err := h.PipelineService.CreatePipeline(&pipeline); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pipeline"})
		return
	}

	c.JSON(http.StatusCreated, pipeline)
}

func (h *PipelineHandler) ListPipelines(c *gin.Context) {
	tenantId := c.GetString("tenantId")
	pipelines, err := h.PipelineService.ListPipelines(tenantId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pipelines"})
		return
	}
	c.JSON(http.StatusOK, pipelines)
}
