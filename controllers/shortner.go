package controllers

import (
	"context"

	"crypto/rand"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HarshithRajesh/url_shortner/initializers"
	"github.com/HarshithRajesh/url_shortner/model"
	"github.com/HarshithRajesh/url_shortner/validators"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var urlHits = make(map[string]*int64)
var mu sync.Mutex
var base62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func base62Encoder(num int) string {
	base := ""
	for num > 0 {
		rem := num % 62
		base = string(base62[rem]) + base
		num = num / 62
	}
	return base
}

// func base62Decoder(s string) int {
// 	number := 0

// 	for _, char := range s {
// 		index := strings.IndexRune(base62, char)
// 		fmt.Println(index)
// 		fmt.Println(number)
// 		number = number*62 + index

// 		fmt.Println(number)
// 	}
// 	return number
// }

func FlushDB() {
	for {
		time.Sleep(5 * time.Second)
		mu.Lock()
		for shorturl, count := range urlHits {
			hits := atomic.SwapInt64(count, 0)
			if hits > 0 {
				var newcount int64
				err := initializers.DB.Raw("UPDATE urls SET hit_count = hit_count + ? WHERE short_url = ? RETURNING hit_count;", hits, shorturl).Scan(&newcount).Error
				if err != nil {
					log.Println("Failed to update the database")
				} else {
					log.Printf("Updated %s and %d with new count:%d\n", shorturl, hits, newcount)
				}
			}
		}
		mu.Unlock()

		keys, _ := initializers.RedisClient.Keys(context.Background(), "hitcount:*").Result()
		for _, key := range keys {
			shortUrl := strings.TrimPrefix(key, "hitcount:")
			hits, _ := initializers.RedisClient.Get(context.Background(), key).Int64()
			if hits > 0 {
				err := initializers.DB.Exec("UPDATE urls SET hit_count = hit_count + ? WHERE short_url=?", hits, shortUrl).Error
				if err == nil {
					initializers.RedisClient.Set(context.Background(), key, 0, 0)
				}
			}
		}
	}

}
func UrlShortner(c *gin.Context) {
	var urladdr model.UrlInput

	if err := c.ShouldBindJSON(&urladdr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validators.ValidateURL(urladdr.Url); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid URL: " + err.Error(),
		})
		return
	}
	if err := validators.ValidateCustomCode(urladdr.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid custom code: " + err.Error(),
		})
		return
	}

	// Use database transaction for atomic operations
	tx := initializers.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check for existing URL within transaction
	var existing model.Urls
	if tx.Where("long_url = ?", urladdr.Url).First(&existing).Error == nil {
		tx.Rollback()
		c.JSON(http.StatusOK, gin.H{"message": existing})
		return
	}

	// Create new URL entry
	url := model.Urls{
		LongUrl: urladdr.Url,
	}

	if err := tx.Create(&url).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create URL",
		})
		return
	}

	// Handle short URL assignment atomically
	if urladdr.Code != "" && strings.TrimSpace(urladdr.Code) != "" {
		// Check if custom code is available within transaction
		var count int64
		tx.Model(&model.Urls{}).Where("short_url = ?", urladdr.Code).Count(&count)
		if count > 0 {
			url.ShortUrl = GetorGenerateRandomUrl(int(url.Id))
		} else {
			url.ShortUrl = urladdr.Code
		}
	} else {
		url.ShortUrl = GetorGenerateRandomUrl(int(url.Id))
	}

	// Update with short URL
	if err := tx.Model(&url).Update("short_url", url.ShortUrl).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update short URL",
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": url,
	})
}

func GetorGenerateRandomUrl(id int) string {
	// Start with base62 encoding of ID
	baseUrl := base62Encoder(id)

	// Try the base URL first
	var count int64
	initializers.DB.Model(&model.Urls{}).Where("short_url = ?", baseUrl).Count(&count)
	if count == 0 {
		return baseUrl
	}

	// If base URL exists, add random suffix
	for attempts := 0; attempts < 10; attempts++ {
		// Generate random suffix
		suffix := generateRandomSuffix(2) // 2 character suffix
		shortUrl := baseUrl + suffix

		initializers.DB.Model(&model.Urls{}).Where("short_url = ?", shortUrl).Count(&count)
		if count == 0 {
			return shortUrl
		}
	}

	// Fallback: use timestamp + random
	timestamp := time.Now().UnixNano()
	randomPart := generateRandomSuffix(3)
	return base62Encoder(int(timestamp%999999)) + randomPart
}

func generateRandomSuffix(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(62))
		result += string(base62[randomIndex.Int64()])
	}
	return result
}

func RedirectUrl(c *gin.Context) {
	shortUrl := c.Param("shortUrl")
	ctx := context.Background()

	useRedis := os.Getenv("USE_REDIS") != "false"

	if useRedis {
		longUrl, err := initializers.RedisClient.Get(ctx, shortUrl).Result()
		if err == nil {
			log.Println("[CACHE HIT] Redis fetched:", shortUrl)
			initializers.RedisClient.Incr(ctx, "hitcount:"+shortUrl)
			c.Redirect(http.StatusFound, longUrl)
			return
		}
		log.Println("[CACHE MISS] Redis miss for:", shortUrl)
	}

	var url model.Urls
	if err := initializers.DB.Where("short_url = ?", shortUrl).First(&url).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	if !useRedis {
		mu.Lock()
		if _, exists := urlHits[shortUrl]; !exists {
			var counts int64 = 0
			urlHits[shortUrl] = &counts
		}
		atomic.AddInt64(urlHits[shortUrl], 1)
		mu.Unlock()
	}

	if useRedis {
		err := initializers.RedisClient.Set(ctx, shortUrl, url.LongUrl, 24*time.Hour).Err()
		if err != nil {
			log.Println("Redis SET failed:", err)
		} else {
			log.Println("Redis SET:", shortUrl, "->", url.LongUrl)
		}
	}

	c.Redirect(http.StatusFound, url.LongUrl)
}
