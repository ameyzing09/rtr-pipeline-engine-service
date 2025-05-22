package main

import (
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/db"
	"github.com/ameyzing09/rtr-pipeline-engine-service/internal/handler"
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

	routes.RegisterPipelineRoutes(api, pipelineHandler)
	router.Run(":8080")
}
