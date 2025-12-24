package tweet

import (
	"context"
	"errors"
)

const MaxTweetLength = 280

var (
	ErrTextTooLong = errors.New("text too long")
)

type TweetService interface {
	Post(ctx context.Context, tweets []Tweet) error
	Get(ctx context.Context, tweetID int64) (*Tweet, error)

	GetFromUsers(ctx context.Context, userIds []int64) ([]Tweet, error)
}

type tweetService struct {
	repo TweetRepo
}

func NewService(ctx context.Context, repo TweetRepo) *tweetService {
	return &tweetService{repo: repo}
}

func (s *tweetService) Post(ctx context.Context, tweets []Tweet) error {
	return s.repo.InsertMany(ctx, tweets)
}

func (s *tweetService) Get(ctx context.Context, tweetID int64) (*Tweet, error) {
	return s.repo.GetTweet(ctx, tweetID)
}

func (s *tweetService) GetFromUsers(ctx context.Context, userIds []int64) ([]Tweet, error) {
	return s.repo.GetTweetsFromUsers(ctx, userIds)	
}