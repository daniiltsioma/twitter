package tweet

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (s *mockTweetService) PostTweet(ctx context.Context, userId int64, text string) (*Tweet, error) {
	tweet := Tweet{
		ID: int64(len(s.tweets)),
		UserID: userId,
		Text: text,
	}

	s.tweets[tweet.ID] = &tweet
	return &tweet, nil
}

func (s *mockTweetService) GetTweet(ctx context.Context, tweetID int64) (*Tweet, error) {
	tweet, ok := s.tweets[tweetID]
	if !ok {
		return nil, errors.New("tweet not found")
	}
	return tweet, nil
}

func (s *mockTweetService) GetTweetsFromUsers(ctx context.Context, usedIds []int64) ([]Tweet, error) {
	return nil, nil
}

func TestHandlerPostTweet(t *testing.T) {
	svc := NewMockTweetService()
	handler := NewHandler(svc)

	tests := []struct{
		name string
		method string
		url string
		body string
		expectedStatus int
	}{
		{"PostTweet_InvalidJSON", http.MethodPost, "/tweet", "{invalid json}", http.StatusBadRequest},
		{"PostTweet_MissingText", http.MethodPost, "/tweet", `{"field": "value"}`, http.StatusBadRequest},
		{"PostTweet_Success", http.MethodPost, "/tweet", `{"text": "hello"}`, http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.PostTweet(rr, req)
			
			if rr.Code != tt.expectedStatus {
				t.Errorf("wrong response code, got %v want %v; %v", rr.Code, tt.expectedStatus, rr.Body)
			}
		})
	}
}

func TestHandlerGetTweet(t *testing.T) {
	svc := NewMockTweetService()
	svc.tweets[1] = &Tweet{ID: 1, Text: "hello"}

	handler := NewHandler(svc)

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