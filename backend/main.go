package main

import (
	"backend/routes"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
	}

	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r)

	// Start the server
	r.Run(":8080") // listen and serve on localhost:8080
}
