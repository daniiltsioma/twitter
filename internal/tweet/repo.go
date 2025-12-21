package tweet

import (
	"context"
	"log"

	"gorm.io/gorm"
)

type TweetRepo interface {
	InsertTweet(ctx context.Context, tweet *Tweet) error
	GetTweet(ctx context.Context, tweetID int64) (*Tweet, error)
}

type tweetRepo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *tweetRepo {
	return &tweetRepo{db: db}
}

func (r *tweetRepo) InsertTweet(ctx context.Context, tweet *Tweet) error {
	if err := gorm.G[Tweet](r.db, gorm.WithResult()).Create(ctx, tweet); err != nil {
		log.Printf("could not insert tweet for userId=%d: %v", tweet.UserID, err)
		return err
	}
	return nil
}

func (r *tweetRepo) GetTweet(ctx context.Context, tweetID int64) (*Tweet, error) {
	tweet, err := gorm.G[Tweet](r.db).Where("id = ?", tweetID).First(ctx)
	if err != nil {
		log.Printf("could not get tweet with ID %d: %v", tweetID, err)
		return nil, err
	}

	return &tweet, err
}