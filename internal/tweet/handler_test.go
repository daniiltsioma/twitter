package tweet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

type mockTweetService struct {
	tweets map[int64]*Tweet
}

func NewMockTweetService() *mockTweetService {
	return &mockTweetService{
		tweets: map[int64]*Tweet{},
	}
}

func (s *mockTweetService) Post(ctx context.Context, tweets []Tweet) error {
	return nil
}

func (s *mockTweetService) Get(ctx context.Context, tweetID int64) (*Tweet, error) {
	tweet, ok := s.tweets[tweetID]
	if !ok {
		return nil, errors.New("tweet not found")
	}
	return tweet, nil
}

func (s *mockTweetService) GetFromUsers(ctx context.Context, usedIds []int64) ([]Tweet, error) {
	return nil, nil
}

func TestHandlerGetTweet(t *testing.T) {
	svc := NewMockTweetService()
	svc.tweets[1] = &Tweet{ID: 1, Text: "hello"}

	handler := NewHandler(context.Background(), svc)

	tests := []struct{
		name string
		tweetID string
		expectedStatus int
	}{
		{"GetTweet_ExistingID", "1", http.StatusOK},
		{"GetTweet_InvalidURL", "hi", http.StatusBadRequest},
		{"GetTweet_NonExistingID", "2", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			// passing tweet id into URL params
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("tweetID", tt.tweetID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			rr := httptest.NewRecorder()

			handler.GetTweet(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("wrong response code, got %v want %v; %v", rr.Code, tt.expectedStatus, rr.Body)
			}
		})
	}
}