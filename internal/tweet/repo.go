package tweet

import (
	"context"
	"log"

	"gorm.io/gorm"
)

type TweetRepo interface {
	InsertTweet(ctx context.Context, tweet *Tweet) error
	InsertMany(ctx context.Context, tweets []Tweet) error
	GetTweet(ctx context.Context, tweetID int64) (*Tweet, error)

	GetTweetsFromUsers(ctx context.Context, userIds []int64) ([]Tweet, error)
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

func (r *tweetRepo) InsertMany(ctx context.Context, tweets []Tweet) error {
	if err := gorm.G[Tweet](r.db, gorm.WithResult()).CreateInBatches(ctx, &tweets, len(tweets)); err != nil {
		log.Printf("could not batch insert tweets: %v", err)
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

func (r *tweetRepo) GetTweetsFromUsers(ctx context.Context, userIds []int64) ([]Tweet, error) {
	tweets, err := gorm.G[Tweet](r.db).Where("user_id IN ?", userIds).Order("created_at DESC").Limit(50).Find(ctx)
	if err != nil {
		log.Printf("could not fetch tweets from users: %v", err)
		return nil, err
	}

	return tweets, err
}