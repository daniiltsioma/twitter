package timeline

import (
	"context"
	"log"

	"github.com/daniiltsioma/twitter/internal/tweet"
	"github.com/daniiltsioma/twitter/internal/user"
)

type TimelineService interface {
	GetTweets(ctx context.Context, userId int64) ([]tweet.Tweet, error)
}

type timelineService struct {
	tweets tweet.TweetService
	users user.UserService
}

func NewService(ts tweet.TweetService, us user.UserService) *timelineService {
	return &timelineService{
		tweets: ts,
		users: us,
	}
}

func (s *timelineService) GetTweets(ctx context.Context, userId int64) ([]tweet.Tweet, error) {
	follows, err := s.users.GetFollows(ctx, userId)
	if err != nil {
		log.Printf("users error: %v", err)
		return nil, err
	}

	userIds := []int64{}
	for _, f := range follows {
		userIds = append(userIds, f.FollowedID)
	}

	tweets, err := s.tweets.GetTweetsFromUsers(ctx, userIds)
	if err != nil {
		log.Printf("tweets error: %v", err)
		return nil, err
	}

	return tweets, err
}