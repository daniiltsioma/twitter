package tweet

import (
	"context"

	"github.com/go-chi/jwtauth"
)

type TweetService interface {
	PostTweet(ctx context.Context, text string) (*Tweet, error)
}

type tweetService struct {
	repo TweetRepo
}

func NewService(repo TweetRepo) *tweetService {
	return &tweetService{repo: repo}
}

func (s *tweetService) PostTweet(ctx context.Context, text string) (*Tweet, error) {
	userId := s.getUserIdFromCtx(ctx)

	tweet := Tweet{
		UserID: userId,
		Text: text,
	}

	if err := s.repo.InsertTweet(ctx, &tweet); err != nil {
		return nil, err
	}

	return &tweet, nil
}

func (s *tweetService) getUserIdFromCtx(ctx context.Context) int64 {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))
	return userID
}