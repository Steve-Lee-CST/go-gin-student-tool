package gin_tool

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeaderKey = "X-Request-ID"

type RequestIDTool struct{}

func (t RequestIDTool) Middleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure the request ID is set in the context
		requestID, exists := GetRequestID(c)
		if !exists || requestID == "" {
			requestID = GenerateRequestID(serviceName)
			c.Request.Header.Set(RequestIDHeaderKey, requestID)
		}
		// Set the request ID in the response header
		c.Writer.Header().Set(RequestIDHeaderKey, requestID)
	}
}

func (t RequestIDTool) Handler(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GenerateRequestID(serviceName)
		c.JSON(http.StatusOK, SuccessResponse(&requestID))
	}
}

func GenerateRequestID(serviceName string) string {
	timestamp, micro := GetCurrentTimestampWithMicro()
	return strings.Join(
		[]string{
			serviceName,
			strconv.FormatInt(timestamp, 10),
			strconv.FormatInt(micro, 10),
			strings.Split(uuid.New().String(), "-")[0],
		},
		":",
	)
}

func GetRequestID(c *gin.Context) (string, bool) {
	requestID, exists := c.Request.Header[http.CanonicalHeaderKey(RequestIDHeaderKey)]
	if !exists || len(requestID) == 0 {
		return "", false
	}
	return requestID[0], true
}
