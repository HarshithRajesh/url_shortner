package main

import (
	"fmt"
	"github.com/HarshithRajesh/url_shortner/controllers"
	"github.com/HarshithRajesh/url_shortner/initializers"
	"github.com/gin-gonic/gin"
	"os"
)

func init() {
	initializers.LoadEnvs()
	initializers.ConnectDB()
	initializers.ConnectRedis()
}

func main() {
	initializers.ConnectRedis()
	if initializers.RedisClient == nil {
		fmt.Println("Redis client is not initialized")
		// continue
	}

	go controllers.FlushDB()
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", controllers.HealthCheck)
	r.GET("/", controllers.Home)
	r.POST("/url", controllers.UrlShortner)
	r.GET("/:shortUrl", controllers.RedirectUrl)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
