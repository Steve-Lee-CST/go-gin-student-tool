package main

import (
	"fmt"

	"github.com/Steve-Lee-CST/go-gin-student-tool/config"
	"github.com/Steve-Lee-CST/go-gin-student-tool/router"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	router.RegisterRouter(engine)

	// engine.SetTrustedProxies([]string{"127.0.0.1"})
	engine.Run(fmt.Sprintf(
		"%s://%s:%s",
		config.Base.Protocol, config.Base.Domain, config.Base.Port,
	))
}
