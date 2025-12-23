package auth

type Credentials struct {
	UserID int64 `gorm:"primaryKey"`
	PasswordHash string
}