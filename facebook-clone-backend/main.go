package main

import (
	"log"
	"os"

	"facebookapi/config"
	"facebookapi/helpers"
	"facebookapi/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to MongoDB first
	config.ConnectDB() // sets global Client

	// Generate JWT key
	key := config.GenerateRandomKey()
	helpers.SetJWTKey(key)

	// Initialize Gin
	r := gin.Default()
	r.Use(cors.Default())

	// Setup routes
	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
