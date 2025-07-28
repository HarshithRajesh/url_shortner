package model

// import "time"

type UrlInput struct {
	Url  string `json:"url" binding:"required"`
	Code string `json:"code"`
}

type Urls struct {
	Id       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	LongUrl  string `gorm:"not null" json:"long_url"`
	ShortUrl string `gorm:"not null;unique" json:"short_url"`
	HitCount int    `json:"hit_count"`
	// CreatedAt time.Time `json:"created_at"`
	// UpdatedAt time.Time `json:"updated_at"`
}
