package routes

import (
	"github.com/ameyzing09/rtr-pipeline-engine-service/handlers"
	"github.com/ameyzing09/rtr-pipeline-engine-service/middleware"
	"github.com/ameyzing09/rtr-pipeline-engine-service/models"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all application routes
func RegisterRoutes(r *gin.Engine, pipelineHandler *handlers.PipelineHandler) {
	// Health check endpoints (no auth required)
	healthHandler := handlers.NewHealthHandler()
	r.GET("/", healthHandler.Health)
	r.GET("/health", healthHandler.Health)
	r.GET("/health/detailed", healthHandler.HealthDetailed)

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

		// GET /pipeline/assignment - Get pipeline assignment by job_id
		pipeline.GET("/assignment",
			middleware.AllowReadOnly(),
			pipelineHandler.GetPipelineAssignment)

		// POST /pipeline/assign - Assign pipeline (ADMIN/HR only)
		pipeline.POST("/assign",
			middleware.RequireRoles(models.RoleAdmin, models.RoleHR),
			pipelineHandler.AssignPipeline)

		// POST /pipeline/seed-defaults - Seed default pipeline templates
		// Self-seed: ADMIN/HR can seed their own tenant (X-Tenant-Id == JWT.tid)
		// Cross-tenant: SUPERADMIN can seed any tenant (X-Tenant-Id = any valid tenant)
		// Middleware order: fail-fast from least to most expensive
		pipeline.POST("/seed-defaults",
			middleware.RequirePipelineDefaultsEnabled(),  // 1. Feature flag (fastest)
			middleware.RequireRoles(models.RoleAdmin, models.RoleHR, models.RoleSuperAdmin), // 2. Role check
			middleware.CrossTenantMiddleware(),           // 3. Extract target tenant
			middleware.RequireTenantActive(),             // 4. Validate tenant state (may call external service)
			pipelineHandler.SeedDefaults)
	}
}
