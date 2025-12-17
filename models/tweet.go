package models

import "time"

type Tweet struct {
	ID int64 `gorm:"primaryKey"`
	UserID int64 `json:"userId"`
	User User `gorm:"foreignKey:UserID"`
	Text string `json:"text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type User struct {
	ID int64 `gorm:"primaryKey"`
	Username string `json:"username" gorm:"uniqueIndex"`
	PasswordHash string `json:"-"`
} 

type Follow struct {
	ID int64 
	FollowerID int64 `json:"followerId" gorm:"primaryKey"`
	FollowedID int64 `json:"followedId" gorm:"primaryKey"`
	Follower User `gorm:"foreignKey:FollowerID"`
	Followed User `gorm:"foreignKey:FollowedID"`
}