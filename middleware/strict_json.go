package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ameyzing09/rtr-pipeline-engine-service/utils"
	"github.com/gin-gonic/gin"
)

// StrictJSONMiddleware ensures that only known fields are accepted in JSON requests
// Rejects requests with unknown fields (Layer 4: ValidationPipe behavior)
// Returns 400 Bad Request if unknown fields are detected
func StrictJSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to POST, PUT, PATCH requests with JSON content
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" {
			c.Next()
			return
		}

		// Check if content type is JSON
		if c.ContentType() != "application/json" {
			c.Next()
			return
		}

		// Read the body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			utils.Error("[StrictJSON] Failed to read request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			c.Abort()
			return
		}

		// Restore body for later reading
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Try to unmarshal into a generic map to detect unknown fields
		var rawData map[string]interface{}
		decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
		decoder.DisallowUnknownFields() // This makes it strict

		if err := decoder.Decode(&rawData); err != nil {
			// Extract the field name from the error if possible
			var errMsg string
			if syntaxErr, ok := err.(*json.SyntaxError); ok {
				errMsg = fmt.Sprintf("Invalid JSON at position %d", syntaxErr.Offset)
			} else {
				errMsg = err.Error()
			}

			utils.Debug("[StrictJSON] Invalid JSON: %s", errMsg)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request format: " + errMsg,
			})
			c.Abort()
			return
		}

		// Continue to next handler
		c.Next()
	}
}
