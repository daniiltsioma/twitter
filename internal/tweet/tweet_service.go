package tweetservice

import (
	"context"

	"github.com/daniiltsioma/twitter/internal/models"
	"github.com/go-chi/jwtauth"
	"gorm.io/gorm"
)

type TweetService interface {
	PostTweet(ctx context.Context, text string) (*models.Tweet, error)
	GetTimeline(ctx context.Context) ([]models.Tweet, error) // this might need improvement
}

type tweetService struct {
	repo *gorm.DB
}

func NewTweetService(repo *gorm.DB) *tweetService {
	return &tweetService{repo: repo}
}

func (s *tweetService) PostTweet(ctx context.Context, text string) (*models.Tweet, error) {
	userId := s.getUserIdFromCtx(ctx)

	tweet := models.Tweet{
		UserID: userId,
		Text: text,
	}

	if err := gorm.G[models.Tweet](s.repo, gorm.WithResult()).Create(ctx, &tweet); err != nil {
		return nil, err
	}

	return &tweet, nil
}

// needs improvement
func (s *tweetService) GetTimeline(ctx context.Context) ([]models.Tweet, error) {
	userId := s.getUserIdFromCtx(ctx)

	var tweets []models.Tweet

	s.repo.Joins("JOIN follows on follows.followed_id = tweets.user_id").Where("follows.follower_id = ?", userId).Order("created_at DESC").Find(&tweets)

	return tweets, nil
}

func (s *tweetService) getUserIdFromCtx(ctx context.Context) int64 {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))
	return userID
}