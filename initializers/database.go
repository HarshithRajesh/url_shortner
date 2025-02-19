package initializers

import (
  "log"
  "os"
  "gorm.io/driver/postgres"
  "gorm.io/gorm"
  "github.com/redis/go-redis/v9"
  "context"
  // "github.com/HarshithRajesh/url_shortner/initializers"
)

var DB *gorm.DB

func ConnectDB(){
  dsn := os.Getenv("DATABASE_URL")
  if dsn == ""{
    log.Fatal("The database url string is empty or not fetched")
  }
  log.Println("Connect the database")

  var err error 
  DB,err = gorm.Open(postgres.Open(dsn),&gorm.Config{})
  if err != nil{
    log.Fatalf("Failed to connect to the database: %v",err)
  }
  log.Println("Successfully connected to the database")


}

var RedisClient *redis.Client 

func ConnectRedis(){
  log.Println("Initializing the redis connection")
  RedisClient = redis.NewClient(&redis.Options{
    Addr : "localhost:6379",
    Password : "",
    DB :  0,
  })

  ctx := context.Background()
  _,err := RedisClient.Ping(ctx).Result()
  if err != nil{
    log.Fatalf("Failed to connect to the Redis Database")
  } else {
    log.Println("connected to Redis")
  }
}

func ConnectDBTest(){
  dsn := os.Getenv("DATABASE_URL_TEST")
  if dsn == ""{
    log.Fatal("The test database url is not fetched")
  }
  log.Println("Connecting to the test database ...")

  var err error
  DB,err = gorm.Open(postgres.Open(dsn),&gorm.Config{})
  if err != nil{
    log.Fatal("Failed to connect to the test database: %v",err)
  }
  log.Println("Successfully connected to the test database")
}
