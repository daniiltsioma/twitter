package handler

import (
	"encoding/json"
	"net/http"

	tweetservice "github.com/daniiltsioma/twitter/internal/tweet"
)

type TweetHandler struct {
	svc tweetservice.TweetService
}

func NewTweetHandler(svc tweetservice.TweetService) *TweetHandler {
	return &TweetHandler{svc: svc}
}

func (h *TweetHandler) PostTweet(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON: " + err.Error(), http.StatusBadRequest)
		return
	}

	if in.Text == "" {
		http.Error(w, "tweet.Text cannot be empty", http.StatusBadRequest)
		return
	}

	tweet, err := h.svc.PostTweet(r.Context(), in.Text)
	if err != nil {
		http.Error(w, "could not post tweet", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tweet)
}

func (h *TweetHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	tweets, err := h.svc.GetTimeline(r.Context())
	if err != nil {
		http.Error(w, "could not get timeline", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweets)
}
