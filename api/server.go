package api

import (
	"link-shortener-api/api/helpers"
	"log"

	"link-shortener-api/api/middlewares"
	"link-shortener-api/api/routes"

	"github.com/gin-gonic/gin"
)

func Init() {
	r := router()
	err := r.Run(helpers.EnvVar("API_PORT", ""))
	if err != nil {
		log.Panic(err)
	}
}

func router() *gin.Engine {
	r := gin.Default()
	r.Use(middlewares.RequestLoggerMiddleware())

	routes.InitializeRoutes(r)
	return r
}
