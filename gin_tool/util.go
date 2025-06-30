package gin_tool

import (
	"time"

	"github.com/gin-gonic/gin"
)

type HandlerWithMiddleware struct {
	Handler    gin.HandlerFunc
	Middleware []gin.HandlerFunc
}

func (h HandlerWithMiddleware) ToChain() gin.HandlersChain {
	chain := make(gin.HandlersChain, 0, len(h.Middleware)+1)
	chain = append(chain, h.Middleware...)
	chain = append(chain, h.Handler)
	return chain
}

func GetCurrentTimestampWithMicro() (int64, int64) {
	fullTime := time.Now().UnixMicro()
	timestamp := fullTime / 1000000
	nano := fullTime % 1000000
	return timestamp, nano
}
