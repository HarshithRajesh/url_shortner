package controllers

import (
  "context"
  "github.com/gin-gonic/gin"
  "github.com/HarshithRajesh/url_shortner/initializers"
  "net/http"
  "time"
)

func Home(c *gin.Context){
  c.JSON(200,gin.H{"message":"Hello"})
}

func HealthCheck(c *gin.Context) {
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  health := gin.H{
    "status": "healthy",
    "timestamp": time.Now().Unix(),
    "services": gin.H{},
  }

  // Check database connection
  sqlDB, err := initializers.DB.DB()
  if err != nil || sqlDB.PingContext(ctx) != nil {
    health["services"].(gin.H)["database"] = "unhealthy"
    health["status"] = "degraded"
  } else {
    health["services"].(gin.H)["database"] = "healthy"
  }

  // Check Redis connection if enabled
  if initializers.RedisClient != nil {
    _, err := initializers.RedisClient.Ping(ctx).Result()
    if err != nil {
      health["services"].(gin.H)["redis"] = "unhealthy"
      health["status"] = "degraded"
    } else {
      health["services"].(gin.H)["redis"] = "healthy"
    }
  }

  status := http.StatusOK
  if health["status"] == "degraded" {
    status = http.StatusServiceUnavailable
  }

  c.JSON(status, health)
}

