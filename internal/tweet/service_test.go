package tweet

import (
	"context"
	"errors"
	"testing"

	"github.com/daniiltsioma/twitter/auth"
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

func (r *mockRepo) GetTweet(ctx context.Context, tweetID int64) (*Tweet, error) {
	tweet, ok := r.tweets[tweetID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &tweet, nil
}

func TestServicePostTweet(t *testing.T) {
	repo := NewMockRepo()
	srv := NewService(repo)

	tests := []struct{
		name string
		text string
		expectedError error
	}{
		{"error if text too long", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus porttitor tellus tempor rhoncus tempus. Phasellus vitae semper velit. Nullam sollicitudin, turpis in porta pellentesque, nisi erat tincidunt dolor, a maximus mauris tortor vestibulum est. Nullam vel risus at velit lobortis efficitur.", ErrTextTooLong},
		{"post tweet successfully", "hello!!!", nil},
	}

	ctx := auth.WithUserID(context.Background(), 123)

	for _, tt := range tests {
		_, err := srv.PostTweet(ctx, tt.text)
		if !errors.Is(err, tt.expectedError) {
			t.Errorf("expected error %v got %v", tt.expectedError, err)
		}
	}
}

func TestServiceGetTweet(t *testing.T) {
	repo := &mockRepo{
		tweets: map[int64]Tweet{
			1: {ID: 1, UserID: 2, Text: "hello"},
		},
	}
	srv := NewService(repo)

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
		_, err := srv.GetTweet(ctx, tt.tweetID)
		if !errors.Is(err, tt.expectedError) {
			t.Errorf("expected error %v got %v", tt.expectedError, err)
		}
	}
}