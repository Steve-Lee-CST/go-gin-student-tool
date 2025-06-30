package router

import (
	"github.com/Steve-Lee-CST/go-gin-student-tool/config"
	"github.com/Steve-Lee-CST/go-gin-student-tool/gin_tool"
	"github.com/Steve-Lee-CST/go-gin-student-tool/handler"
	"github.com/gin-gonic/gin"
)

func toolRouter(engine *gin.RouterGroup) {
	// tool/request_id: GET Request ID
	engine.GET("request_id", handler.RequestIDHandler)
	// tool/ping: GET/POST Ping: response is the decoded-request
	engine.GET("ping", gin_tool.HandlerWithMiddleware{
		Handler: handler.PingHandler,
		Middleware: []gin.HandlerFunc{
			// Request ID middleware
			gin_tool.RequestIDTool{}.Middleware(config.Base.ServiceName),
			// HTTP Logger middleware
			gin_tool.HttpLoggerTool{}.Middleware(gin_tool.DefaultHttpLogger),
			// HTTP helper middleware
			gin_tool.HttpHelper{}.Middleware(),
		},
	}.ToChain()...)
	engine.POST("ping", gin_tool.HandlerWithMiddleware{
		Handler: handler.PingHandler,
		Middleware: []gin.HandlerFunc{
			// Request ID middleware
			gin_tool.RequestIDTool{}.Middleware(config.Base.ServiceName),
			// HTTP helper middleware
			gin_tool.HttpHelper{}.Middleware(),
		},
	}.ToChain()...)
}
