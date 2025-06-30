package handler

import (
	"github.com/Steve-Lee-CST/go-gin-student-tool/config"
	"github.com/Steve-Lee-CST/go-gin-student-tool/gin_tool"
	"github.com/gin-gonic/gin"
)

func RequestIDHandler(c *gin.Context) {
	gin_tool.RequestIDTool{}.Handler(config.Base.ServiceName)(c)
}
