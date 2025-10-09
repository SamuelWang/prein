package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
)

// Logger is a middleware function that logs the incoming requests.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Request: %s %s", c.Request.Method, c.Request.URL)
		c.Next()
	}
}

// Recovery is a middleware function that recovers from panics and writes a 500 if there was one.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Recovered from panic:", err)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}