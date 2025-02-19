package controllers

import (
  "fmt"
  "strings"
  "github.com/gin-gonic/gin"
  "github.com/HarshithRajesh/url_shortner/model"
  "github.com/HarshithRajesh/url_shortner/initializers"
  "net/http"
  "gorm.io/gorm"
  "sync/atomic"
  "time"
  "log"
  "sync"
  "context"
)
var urlHits = make(map[string]*int64)
var mu sync.Mutex
var base62 ="0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
func base62Encoder(num int)string{
  base:=""
  for num>0{
    rem:=num%62
    base= string(base62[rem])+base
    num=num/62
  }
  return base
}

func base62Decoder(s string)int{
  number:=0

  for _ , char := range s {
    index:=strings.IndexRune(base62,char)
    fmt.Println(index)
    fmt.Println(number)
    number = number*62+index

    fmt.Println(number)
  }
  return number
}

func FlushDB(){
  for{
    time.Sleep(5*time.Second)
       mu.Lock()    
    for shorturl,count := range urlHits{
      hits := atomic.SwapInt64(count,0)
      if hits>0{
        var newcount int64
        err := initializers.DB.Raw("UPDATE urls SET hit_count = hit_count + ? WHERE short_url = ? RETURNING hit_count;",hits,shorturl).Scan(&newcount).Error 
        if err !=nil{
          log.Println("Failed to update the database")
        } else{
          log.Printf("Updated %s and %d with new count:%d\n",shorturl,hits,newcount)
        }
      }
    }
    mu.Unlock()

    keys,_ := initializers.RedisClient.Keys(context.Background(),"hitcount:*").Result()
    for _, key := range keys{
      shortUrl := strings.TrimPrefix(key,"hitcount:")
      hits,_ := initializers.RedisClient.Get(context.Background(),key).Int64()
      if hits > 0{
        err := initializers.DB.Exec("UPDATE urls SET hit_count = hit_count + ? WHERE short_url=?",hits,shortUrl).Error
        if err == nil{
          initializers.RedisClient.Set(context.Background(),key,0,0)
        }
      }
    }
  }

}
func UrlShortner(c *gin.Context){
  var urladdr model.UrlInput

  if err:= c.ShouldBindJSON(&urladdr);err != nil{
    c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
    return
  }

  var existing model.Urls
  if initializers.DB.Where("long_url = ?",urladdr.Url).First(&existing).Error == nil {
    c.JSON(http.StatusOK,gin.H{"message":existing})
    return
  }
  url := model.Urls{
    LongUrl: urladdr.Url,
  }

  if err := initializers.DB.Create(&url).Error; err != nil{
    c.JSON(http.StatusInternalServerError,gin.H{
      "error":"Failed to create a url",
    })
    return
  }
    if urladdr.Code != " " {
      var count int64
      initializers.DB.Model(&model.Urls{}).Where("short_url=?",urladdr.Code).Count(&count)
      if count > 0{
        url.ShortUrl = GetorGenerateRandomUrl(int(url.Id))
      } else{
        url.ShortUrl = urladdr.Code
      }
    } else {
        url.ShortUrl = GetorGenerateRandomUrl(int(url.Id))
    }
  // url.ShortUrl = base62Encoder(int(url.Id))
  // codeId := base62Decoder(urladdr.Code)
  // url.ShortUrl = GetorGenerateRandomUrl(int(url.Id),codeId,urladdr.Code)
  initializers.DB.Save(&url)

  c.JSON(http.StatusOK,gin.H{
    "message":url,
  })
}


func GetorGenerateRandomUrl(id int) string{
  shortUrl := base62Encoder(id)
  var count int64 
  for{
    initializers.DB.Model(&model.Urls{}).Where("short_url = ?",shortUrl).Count(&count)
    if count == 0{
      break
    }
    id++
    shortUrl = base62Encoder(id)
  }
  return shortUrl
}
func RedirectUrl(c *gin.Context){
  shortUrl := c.Param("shortUrl")
  ctx := context.Background()

  longUrl,err := initializers.RedisClient.Get(ctx,shortUrl).Result()
  if err == nil{
    log.Println("Fetched from redis: ",shortUrl)
    initializers.RedisClient.Incr(ctx,"hitcount:"+shortUrl)
    c.Redirect(http.StatusFound,longUrl)
    return
  }


  log.Println("Fetched from redis: ",shortUrl)
  var url model.Urls

  if err := initializers.DB.Where("short_url=?",shortUrl).First(&url).Error; err != nil{
    if err == gorm.ErrRecordNotFound{
    c.JSON(http.StatusNotFound,gin.H{
      "error":"user not found",
    })
  } else{
      c.JSON(http.StatusInternalServerError,gin.H{"error":"Database error"})
  } 
  return
}
mu.Lock()
if _,exists := urlHits[shortUrl];!exists{
  var counts int64 = 0
  urlHits[shortUrl] = &counts 
} 
  atomic.AddInt64(urlHits[shortUrl],1)

  mu.Unlock()
  log.Println("Setting Redis key:", shortUrl, "->", url.LongUrl)
  err = initializers.RedisClient.Set(ctx,shortUrl,url.LongUrl,24*time.Hour).Err()
  if err != nil{
    log.Println("Failed to set Redis key: ",err)
  }
  c.Redirect(http.StatusFound,url.LongUrl)

}


