package controllers

import (
  "fmt"
  "strings"
  "github.com/gin-gonic/gin"
  "github.com/HarshithRajesh/url_shortner/model"
  "github.com/HarshithRajesh/url_shortner/initializers"
  "net/http"
  "gorm.io/gorm"
)

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
 url.HitCount+=1
 initializers.DB.Save(&url)
 c.Redirect(http.StatusFound,url.LongUrl)

}


