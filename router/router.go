package router

import (
	"github.com/Steve-Lee-CST/go-gin-student-tool/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRouter(engine *gin.Engine) {
	toolRouter(engine.Group("tool"))

	engine.NoRoute(middleware.NotFoundHandler)
}
