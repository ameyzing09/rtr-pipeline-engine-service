package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ameyzing09/rtr-pipeline-engine-service/config"
	"github.com/ameyzing09/rtr-pipeline-engine-service/db"
	"github.com/ameyzing09/rtr-pipeline-engine-service/handlers"
	"github.com/ameyzing09/rtr-pipeline-engine-service/middleware"
	"github.com/ameyzing09/rtr-pipeline-engine-service/repositories"
	"github.com/ameyzing09/rtr-pipeline-engine-service/routes"
	"github.com/ameyzing09/rtr-pipeline-engine-service/services"
	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	// Enforce UTC timezone for all time operations
	time.Local = time.UTC

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set Gin mode
	if cfg.Server.GinMode != "" {
		gin.SetMode(cfg.Server.GinMode)
	}

	// Set log level
	utils.SetLogLevel(cfg.Logging.Level)
	utils.Info("Starting application with environment: %s", cfg.Server.Env)

	// Initialize database
	dbInstance, err := db.InitDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	utils.Info("Database connection established")

	// Initialize repositories
	repos := initRepositories(dbInstance)

	// Initialize services
	svcs := initServices(repos)

	// Initialize handlers
	hndlrs := initHandlers(svcs)

	// Setup router
	router := setupRouter(hndlrs, cfg)

	// Start server
	addr := ":" + cfg.Server.Port
	utils.Info("Server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

type repositoriesContainer struct {
	PipelineRepo repositories.PipelineRepository
}

func initRepositories(db *gorm.DB) *repositoriesContainer {
	return &repositoriesContainer{
		PipelineRepo: repositories.NewPipelineRepository(db),
	}
}

type servicesContainer struct {
	PipelineService services.PipelineService
}

func initServices(repos *repositoriesContainer) *servicesContainer {
	return &servicesContainer{
		PipelineService: services.NewPipelineService(repos.PipelineRepo),
	}
}

type handlersContainer struct {
	PipelineHandler *handlers.PipelineHandler
}

func initHandlers(svcs *servicesContainer) *handlersContainer {
	return &handlersContainer{
		PipelineHandler: handlers.NewPipelineHandler(svcs.PipelineService),
	}
}

func setupRouter(hndlrs *handlersContainer, cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Apply middleware in order (7-layer validation flow):
	// Layer 2: CORS - Allow cross-origin requests from configured origins
	router.Use(middleware.CORSMiddleware(cfg))

	// Layer 3: Rate Limiter - Check first to prevent abuse
	router.Use(middleware.RateLimiterMiddleware(cfg))

	// Layer 4: Strict JSON validation - Reject unknown fields
	router.Use(middleware.StrictJSONMiddleware())

	// Layer 1: JWT validation - Parse token and inject user context
	router.Use(middleware.JWTMiddleware(cfg))

	// Layer 1: Tenant validation - Validate tenant ID header and cross-check with JWT
	router.Use(middleware.TenantMiddleware())

	// Request/Response Logger - Log all requests with tenant_id, user_id, route, status
	router.Use(middleware.LoggerMiddleware())

	// Register routes
	routes.RegisterRoutes(router, hndlrs.PipelineHandler)

	return router
}
