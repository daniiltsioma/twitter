package tweet

import (
	"context"
	"errors"
	"testing"

	"github.com/daniiltsioma/twitter/internal/auth"
	"gorm.io/gorm"
)

type mockRepo struct {
	tweets map[int64]Tweet
}

func NewMockRepo() *mockRepo {
	return &mockRepo{
		tweets: make(map[int64]Tweet),
	}
}

func (r *mockRepo) InsertTweet(ctx context.Context, tweet *Tweet) error {
	tweet.ID = int64(len(r.tweets))
	r.tweets[tweet.ID] = *tweet
	return nil
}

func (r *mockRepo) InsertMany(ctx context.Context, tweets []Tweet) error {
	for _, tweet := range tweets {
		r.tweets[tweet.ID] = tweet
	}
	return nil
}

func (r *mockRepo) GetTweet(ctx context.Context, tweetID int64) (*Tweet, error) {
	tweet, ok := r.tweets[tweetID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &tweet, nil
}

func (r *mockRepo) GetTweetsFromUsers(ctx context.Context, userIds []int64) ([]Tweet, error) {
	return nil, nil
}

func TestServiceGetTweet(t *testing.T) {
	repo := &mockRepo{
		tweets: map[int64]Tweet{
			1: {ID: 1, UserID: 2, Text: "hello"},
		},
	}
	srv := NewService(context.Background(), repo)

	tests := []struct{
		name string
		tweetID int64
		expectedError error
	}{
		{"finds an existing tweet", 1, nil},
		{"does not find a non-existing tweet", 2, gorm.ErrRecordNotFound},
	}

	ctx := auth.WithUserID(context.Background(), 123)

	for _, tt := range tests {
		_, err := srv.Get(ctx, tt.tweetID)
		if !errors.Is(err, tt.expectedError) {
			t.Errorf("expected error %v got %v", tt.expectedError, err)
		}
	}
}