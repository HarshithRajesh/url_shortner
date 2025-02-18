package tests

import (
  "time"
  "os"
  "log"
  "sync"
  "net/http"
  "net/http/httptest"
  "testing"
  "fmt"
  "github.com/gin-gonic/gin"
  "github.com/HarshithRajesh/url_shortner/controllers"
  "github.com/stretchr/testify/assert"
  "github.com/HarshithRajesh/url_shortner/initializers"
  "github.com/HarshithRajesh/url_shortner/model"
  "github.com/joho/godotenv"
)
func TestMain(m *testing.M) {
    // Load the environment variables
    err := godotenv.Load("../.env")
    if err != nil {
        log.Println("Warning: .env file was not found")
    }

    testDBUrl := os.Getenv("DATABASE_URL_TEST")
    fmt.Println("TEST DB URL: ", testDBUrl)

    if testDBUrl == "" {
        log.Fatal("Test database URL is empty")
    }

    // Set the environment variable for the test DB
    os.Setenv("DATABASE_URL_TEST", testDBUrl)

    // Connect to the test database
    initializers.ConnectDBTest()

    // Run migrations on the test DB
    if err := initializers.DB.AutoMigrate(&model.Urls{}); err != nil {
        log.Fatalf("Failed to migrate the test DB: %v", err)
    }else{
      fmt.Println("Migration completed")
    }

    // Run the tests
    code := m.Run()

    // Exit with the test result code
    os.Exit(code)
}

func TestConcurrentHits(t *testing.T){ 
  go controllers.FlushDB()
  gin.SetMode(gin.TestMode)
  router := gin.Default()

  router.GET("/:shortUrl",controllers.RedirectUrl)
  // router.GET("/:shortUrl",controllers.RedirectUrl)

  shortUrl := "neo"
  var existingUrl model.Urls 
  if err := initializers.DB.Where("short_url=?",shortUrl).First(&existingUrl).Error; err !=nil{ 
  initializers.DB.Create(&model.Urls{
    LongUrl: "https://google.com",
    ShortUrl: shortUrl,
    HitCount : 0,
  })
}
  var wg sync.WaitGroup
  concurrentRequests := 10

  for i:=0;i<concurrentRequests;i++{
    wg.Add(1)
    go func(){
      defer wg.Done()
      req , _ := http.NewRequest("GET",fmt.Sprintf("/%s", shortUrl),nil)

      resp := httptest.NewRecorder()
      router.ServeHTTP(resp,req)
      assert.Equal(t,http.StatusFound,resp.Code,"Expected 302 Redirect")
    }()
  }
  wg.Wait()
time.Sleep(6 * time.Second) // More than the 5s interval in FlushDB

    // Fetch the updated hit count from DB
    var url model.Urls
    initializers.DB.First(&url, "short_url = ?", shortUrl)
    fmt.Printf("After FlushDB: HitCount = %d\n", url.HitCount)

    // Check if the hit count matches the number of concurrent requests
    assert.Equal(t, concurrentRequests, url.HitCount, "Hit count should be equal to number of concurrent requests")
}
