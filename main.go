package main
import(
  "fmt"
  "github.com/gin-gonic/gin"
  "github.com/HarshithRajesh/url_shortner/controllers"
  "github.com/HarshithRajesh/url_shortner/initializers"
)
func init(){
  initializers.LoadEnvs()
  initializers.ConnectDB()
  initializers.ConnectRedis()
}

func main(){
  initializers.ConnectRedis()
   if initializers.RedisClient == nil {
	    fmt.Println("Redis client is not initialized")
	    // continue
		}

  go controllers.FlushDB()
  r:= gin.Default()

  r.GET("/",controllers.Home)
  r.POST("/url",controllers.UrlShortner)
  r.GET("/:shortUrl",controllers.RedirectUrl)
  r.Run()
}
