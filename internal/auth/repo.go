package auth

import (
	"context"
	"log"

	"gorm.io/gorm"
)

type AuthRepo interface {
	InsertCredentials(ctx context.Context, userId int64, passwordHash string) error
	GetPasswordHash(ctx context.Context, userId int64) (string, error)
}

type authRepo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *authRepo {
	return &authRepo{db: db}
}

func (r *authRepo) InsertCredentials(ctx context.Context, userId int64, passwordHash string) error {
	credentials := &Credentials{
		UserID: userId,
		PasswordHash: passwordHash,
	}

	err := gorm.G[Credentials](r.db, gorm.WithResult()).Create(ctx, credentials)
	if err != nil {
		log.Printf("error storing credentials for userId=%d: %v", userId, err)
	}

	return err
}

func (r *authRepo) GetPasswordHash(ctx context.Context, userId int64) (string, error) {
	credentials, err := gorm.G[Credentials](r.db).Where("user_id = ?", userId).First(ctx)
	if err != nil {
		log.Printf("error fetching credentials for userId=%d: %v", userId, err)
		return "", err
	}

	return credentials.PasswordHash, err
}