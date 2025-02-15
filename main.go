package main
import(
  "github.com/gin-gonic/gin"
  "github.com/HarshithRajesh/url_shortner/controllers"
  "github.com/HarshithRajesh/url_shortner/initializers"
)
func init(){
  initializers.LoadEnvs()
  initializers.ConnectDB()
}

func main(){
  r:= gin.Default()

  r.GET("/",controllers.Home)
  r.POST("/url",controllers.UrlShortner)
  r.GET("/:shortUrl",controllers.RedirectUrl)
  r.Run()
}
