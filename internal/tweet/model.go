package tweet

import "time"

type Tweet struct {
	ID int64 `gorm:"primaryKey"`
	UserID int64 `json:"userId"`
	Text string `json:"text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}