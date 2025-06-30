package gin_tool

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpLoggerTool struct{}

func NewHttpLoggerTool() *HttpLoggerTool {
	return &HttpLoggerTool{}
}

func (t HttpLoggerTool) Middleware(
	httpLogger func(*HttpRequest, *HttpResponse, int64),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Proceed with the request
		c.Next()

		duration := time.Since(startTime).Microseconds()

		httpRequest, _ := GetHttpRequest(c)
		httpResponse, _ := GetHttpResponse(c)

		if httpLogger != nil {
			httpLogger(httpRequest, httpResponse, duration)
		}
	}
}

func DefaultHttpLogger(httpRequest *HttpRequest, httpResponse *HttpResponse, duration int64) {
	if httpRequest == nil || httpResponse == nil {
		return
	}
	reqBytes, _ := json.Marshal(httpRequest)
	rspBytes, _ := json.Marshal(httpResponse)
	// Print request and response to console
	// In production, you might want to log this to a file or a logging service
	fmt.Printf("HTTP Logger: %s ================\n", httpRequest.RequestID)
	fmt.Printf("Request: %s\n", reqBytes)
	fmt.Printf("Response: %s\n", rspBytes)
	fmt.Printf("Duration: %d μs\n", duration)
	fmt.Printf("HTTP Logger: %s ================\n", httpRequest.RequestID)
}
