package user

import (
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type UserRepo interface {
	InsertUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, username string) (User, error)

	InsertFollow(ctx context.Context, followerId, followedId int64) error
	DeleteFollow(ctx context.Context, followerId, followedId int64) error
}

type userRepo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) InsertUser(ctx context.Context, user *User) error {
	if err := gorm.G[User](r.db, gorm.WithResult()).Create(ctx, user); err != nil {
		log.Printf("failed to insert user %s: %v", user.Username, err)
		return err
	}
	return nil
}

func (r *userRepo) GetUser(ctx context.Context, username string) (User, error) {
	return gorm.G[User](r.db).Where("username = ?", username).First(ctx)
}

func (r *userRepo) InsertFollow(ctx context.Context, followerId, followedId int64) error {
	follow := Follow{
		FollowerID: followerId,
		FollowedID: followedId,
	}

	if err := gorm.G[Follow](r.db, gorm.WithResult()).Create(ctx, &follow); err != nil {
		log.Printf("could not create a follow for followerId=%d, followedId=%d: %v", followerId, followedId, err)
		return err
	}

	return nil
}

func (r *userRepo) DeleteFollow(ctx context.Context, followerId, followedId int64) error {
	n, err := gorm.G[Follow](r.db).Where("follower_id = ? AND followed_id = ?", followerId, followedId).Delete(ctx)
	if err != nil {
		log.Printf("could not delete a follow for followerId=%d, followedId=%d: %v", followerId, followedId, err)
		return err
	}
	if n == 0 {
		return fmt.Errorf("user %d does not follow %d", followerId, followedId)
	}

	return nil
}