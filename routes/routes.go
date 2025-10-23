package routes

import (
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/handlers"
	"github.com/ameyzing09/rtr-pipeline-engine-service/middleware"
	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all application routes
func RegisterRoutes(r *gin.Engine, pipelineHandler *handlers.PipelineHandler) {
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "rtr-pipeline-engine-service: ok")
	})

	// Pipeline routes
	pipeline := r.Group("/pipeline")
	{
		// POST /pipeline - Create pipeline (ADMIN/HR only)
		pipeline.POST("",
			middleware.RequireRoles(models.RoleAdmin, models.RoleHR),
			pipelineHandler.CreatePipeline)

		// GET /pipeline - List pipelines (all roles, INTERVIEWER read-only via middleware)
		pipeline.GET("",
			middleware.AllowReadOnly(),
			pipelineHandler.ListPipelines)

		// GET /pipeline/:id - Get pipeline by ID (all roles, INTERVIEWER read-only via middleware)
		pipeline.GET("/:id",
			middleware.AllowReadOnly(),
			pipelineHandler.GetPipelineByID)

		// PATCH /pipeline/:id - Update pipeline (ADMIN/HR only)
		pipeline.PATCH("/:id",
			middleware.RequireRoles(models.RoleAdmin, models.RoleHR),
			pipelineHandler.UpdatePipeline)

		// POST /pipeline/assign - Assign pipeline (ADMIN/HR only)
		pipeline.POST("/assign",
			middleware.RequireRoles(models.RoleAdmin, models.RoleHR),
			pipelineHandler.AssignPipeline)
	}
}
