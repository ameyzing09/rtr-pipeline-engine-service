package routes

import (
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterPipelineRoutes(r *gin.RouterGroup, h *handler.PipelineHandler) {
	pipeline := r.Group("/pipeline")
	//middleware to be attached here
	// e.g. pipeline.Use(middleware.AuthMiddleware())
	pipeline.POST("/", h.CreatePipeline)
	pipeline.GET("/", h.ListPipelines)
	pipeline.POST("/assign", h.AssignPipeline)
	// future pipeline routes can be added here
}
