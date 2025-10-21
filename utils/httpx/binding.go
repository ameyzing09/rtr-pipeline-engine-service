package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// HandleBindingError handles JSON binding and validation errors
func HandleBindingError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fieldErr := range validationErrs {
			errors[fieldErr.Field()] = getValidationErrorMessage(fieldErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validation failed",
			"fields": errors,
		})
		return
	}

	// Generic binding error
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Invalid request format: " + err.Error(),
	})
}

// getValidationErrorMessage returns a human-readable error message for validation errors
func getValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short (minimum: " + fieldErr.Param() + ")"
	case "max":
		return "Value is too long (maximum: " + fieldErr.Param() + ")"
	case "uuid":
		return "Must be a valid UUID"
	case "oneof":
		return "Must be one of: " + fieldErr.Param()
	default:
		return "Validation failed: " + fieldErr.Tag()
	}
}
