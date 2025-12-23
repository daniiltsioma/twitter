package tweet

import (
	"context"
	"errors"

	"github.com/daniiltsioma/twitter/internal/auth"
)

const MaxTweetLength = 280

var (
	ErrTextTooLong = errors.New("text too long")
)

type TweetService interface {
	PostTweet(ctx context.Context, text string) (*Tweet, error)
	GetTweet(ctx context.Context, tweetID int64) (*Tweet, error)

	GetTweetsFromUsers(ctx context.Context, userIds []int64) ([]Tweet, error)
}

type tweetService struct {
	repo TweetRepo
}

func NewService(repo TweetRepo) *tweetService {
	return &tweetService{repo: repo}
}

func (s *tweetService) PostTweet(ctx context.Context, text string) (*Tweet, error) {
	userId, ok := auth.UserIDFromContext(ctx); 
	if !ok {
		return nil, errors.New("no user id")
	}

	if len(text) > 280 {
		return nil, ErrTextTooLong
	}

	tweet := Tweet{
		UserID: userId,
		Text: text,
	}

	if err := s.repo.InsertTweet(ctx, &tweet); err != nil {
		return nil, err
	}

	return &tweet, nil
}

func (s *tweetService) GetTweet(ctx context.Context, tweetID int64) (*Tweet, error) {
	return s.repo.GetTweet(ctx, tweetID)
}

func (s *tweetService) GetTweetsFromUsers(ctx context.Context, userIds []int64) ([]Tweet, error) {
	return s.repo.GetTweetsFromUsers(ctx, userIds)	
}