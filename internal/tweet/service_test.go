package tweet

import (
	"context"
	"errors"
	"testing"

	"github.com/daniiltsioma/twitter/auth"
)

type mockRepo struct {}

func (r *mockRepo) InsertTweet(ctx context.Context, tweet *Tweet) error {
	tweet.ID = 456
	return nil
}

func TestServicePostTweet(t *testing.T) {
	repo := mockRepo{}
	srv := NewService(&repo)

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
