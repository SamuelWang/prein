package main

import (
	"backend/routes"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
	}

	r := gin.Default()

	// Setup session store
	// Read cookie secret from environment. Use a fallback for local dev only.
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET environment variable not set. Please set SESSION_SECRET to a secure random value.")
	}
	store := cookie.NewStore([]byte(sessionSecret))
	r.Use(sessions.Sessions("cookie_session", store))

	// Setup routes
	routes.SetupRoutes(r)

	// Start the server
	r.Run(":8080") // listen and serve on localhost:8080
}
