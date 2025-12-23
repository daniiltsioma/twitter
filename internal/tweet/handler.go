package tweet

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/daniiltsioma/twitter/internal/auth"
	"github.com/go-chi/chi"
)

type TweetHandler struct {
	svc TweetService
}

func NewHandler(svc TweetService) *TweetHandler {
	return &TweetHandler{svc: svc}
}

func (h *TweetHandler) PostTweet(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return 
	}

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

	tweet, err := h.svc.PostTweet(r.Context(), userId, in.Text)
	if err != nil {
		http.Error(w, "could not post tweet", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tweet)
}

func (h *TweetHandler) GetTweet(w http.ResponseWriter, r *http.Request) {
	tweetID, err := strconv.Atoi(chi.URLParam(r, "tweetID"))
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	tweet, err := h.svc.GetTweet(r.Context(), int64(tweetID))
	if err != nil {
		http.Error(w, "tweet not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweet)
}