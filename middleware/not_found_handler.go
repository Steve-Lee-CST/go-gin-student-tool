package middleware

import "github.com/gin-gonic/gin"

func NotFoundHandler(c *gin.Context) {
	c.JSON(404, gin.H{
		"error":   "Not Found",
		"message": "The requested resource could not be found.",
	})
	c.Abort()
}
