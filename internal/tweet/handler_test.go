package tweet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockTweetService struct {}

func (s *mockTweetService) PostTweet(ctx context.Context, text string) (*Tweet, error) {
	return &Tweet{Text: text}, nil
}

func TestHandlerPostTweet(t *testing.T) {
	svc := &mockTweetService{}
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