package routes

import (
	"facebookapi/controllers"
	"facebookapi/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	// Public routes
	router.POST("/signup", controllers.Signup)
	router.POST("/login", controllers.Login)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.Authenticate())

	{
		// protected.GET("/users", controllers.GetUsers)
		// protected.GET("/user/:id", controllers.GetUser)
	}
}
