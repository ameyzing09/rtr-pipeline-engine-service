package middleware

import (
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/config"
	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
)

// RequireFeatureFlag creates middleware that checks if a feature flag is enabled
// Returns 403 Forbidden if the feature is disabled
func RequireFeatureFlag(flagName string, isEnabled func(*config.Config) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()

		if !isEnabled(cfg) {
			utils.Warn("[FeatureFlag] Feature '%s' is disabled", flagName)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Feature not enabled",
				"feature": flagName,
				"code":    "FEATURE_DISABLED",
			})
			c.Abort()
			return
		}

		utils.Debug("[FeatureFlag] Feature '%s' is enabled", flagName)
		c.Next()
	}
}

// RequirePipelineDefaultsEnabled creates middleware that checks if pipeline defaults seeding is enabled
func RequirePipelineDefaultsEnabled() gin.HandlerFunc {
	return RequireFeatureFlag("PIPELINE_DEFAULTS_ENABLED", func(cfg *config.Config) bool {
		return cfg.PipelineDefaultsEnabled
	})
}
