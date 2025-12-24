package tweet

import "time"

type Tweet struct {
	ID int64 `gorm:"primaryKey"`
	UserID int64 `json:"userId" gorm:"index:idx_user_created,priority:1"`
	Text string `json:"text"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_user_created,priority:2"`
}