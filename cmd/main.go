package main

import (
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/db"
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/handler"
	tenant_middleware "github.com/ameyzing09/rtr-pipeline-engine-service/internal/middleware"
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/routes"
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	dbInstance := db.InitDB()

	pipelineService := service.NewPipelineService(dbInstance)
	pipelineHandler := handler.NewPipelineHandler(pipelineService)

	router := gin.Default()
	api := router.Group("/api")

	api.Use(tenant_middleware.TenantMiddleware())

	routes.RegisterPipelineRoutes(api, pipelineHandler)
	router.Run(":8081")
}
