package handler

import (
	"net/http"

	"github.com/Steve-Lee-CST/go-gin-student-tool/gin_tool"
	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	request, _ := gin_tool.GetHttpRequest(c)
	c.JSON(http.StatusOK, gin_tool.SuccessResponse(request))
}
