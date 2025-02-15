package main

import (
  "github.com/HarshithRajesh/url_shortner/initializers"
  "github.com/HarshithRajesh/url_shorner/model"
  "log"
)

func init(){
  initializers.LoadEnvs()
  initializers.ConnectDB()
}

func main(){
  log.Println("Starting migration")
  err:= initializers.DB.AutoMigrate(
    &model.Urls{},
  )
  if err != nil{
    log.Fatal("Failed to migrate")
  } else {
    log.Fatal("Migration successfull")
  }
}
