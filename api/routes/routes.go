package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"link-shortener-api/api/controllers"
	"link-shortener-api/api/middlewares"
)

func InitializeRoutes(router *gin.Engine) {
	router.GET("/", welcome)
	router.NoRoute(notFound)

	userController := controllers.UserController{}
	linkController := controllers.LinkController{}
	adminController := controllers.AdminController{}

	router.GET("/:url", linkController.ShortURLHandler)
	linksGroup := router.Group("/links")
	linksGroup.Use(middlewares.Authentication())
	linksGroup.POST("/new", linkController.URLShortener)

	u := router.Group("/users")
	{
		u.POST("/signup", userController.SignUp)
		u.POST("/login", userController.Login)
	}

	adminGroup := router.Group("/admin")
	adminGroup.POST("/signup", adminController.SignUp)
	adminGroup.POST("/login", adminController.Login)
	adminGroup.Use(middlewares.AdminAuthentication())

	adminGroup.GET("/users", userController.GetAllUsers)

}

func welcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "Welcome To API",
	})
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"status":  404,
		"message": "Route Not Found",
	})
}
