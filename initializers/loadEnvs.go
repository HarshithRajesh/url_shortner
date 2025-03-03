package initializers


import (
  // "os"
  "log"
  "github.com/joho/godotenv"
)

func LoadEnvs(){
  // err := os.Getenv()
  err := godotenv.Load()
  if err != nil{
    log.Fatal("Error loading the envs")
  }
}

