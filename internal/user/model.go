package user

type User struct {
	ID int64 `gorm:"primaryKey"`
	Username string `json:"username" gorm:"uniqueIndex"`
	PasswordHash string `json:"-"`
} 

type Follow struct {
	ID int64 `gorm:"primaryKey"`
	FollowerID int64 `json:"followerId" gorm:"primaryKey"`
	FollowedID int64 `json:"followedId" gorm:"primaryKey"`
	Follower User `gorm:"foreignKey:FollowerID"`
	Followed User `gorm:"foreignKey:FollowedID"`
}