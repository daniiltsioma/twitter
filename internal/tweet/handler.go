package tweet

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/daniiltsioma/twitter/internal/auth"
	"github.com/go-chi/chi"
)

type TweetHandler struct {
	svc TweetService
	tweetCh chan Tweet
	maxBatchSize int
	maxWait time.Duration
}

func NewHandler(ctx context.Context, svc TweetService) *TweetHandler {
	h := &TweetHandler{
		svc: svc,
		tweetCh: make(chan Tweet, 200),
		maxBatchSize: 300,
		maxWait: 50 * time.Millisecond,
	}

	go h.worker(ctx)
	return h
}

func (h *TweetHandler) PostTweet(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return 
	}

	var in Tweet

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON: " + err.Error(), http.StatusBadRequest)
		return
	}

	if in.Text == "" {
		http.Error(w, "tweet.Text cannot be empty", http.StatusBadRequest)
		return
	}

	in.UserID = userId

	select {
	case h.tweetCh <- in:
		w.WriteHeader(http.StatusAccepted)
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (h *TweetHandler) GetTweet(w http.ResponseWriter, r *http.Request) {
	tweetID, err := strconv.Atoi(chi.URLParam(r, "tweetID"))
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	tweet, err := h.svc.Get(r.Context(), int64(tweetID))
	if err != nil {
		http.Error(w, "tweet not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweet)
}

func (h *TweetHandler) worker(ctx context.Context) {
	ticker := time.NewTicker(h.maxWait)
	defer ticker.Stop()

	batch := make([]Tweet, 0, h.maxBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		_ = h.svc.Post(ctx, batch)
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case t := <-h.tweetCh:
			batch = append(batch, t)
			if len(batch) == h.maxBatchSize {
				log.Printf("batch full")
				flush()
			}
		case <-ticker.C:
			if len(batch) != 0 {
				log.Printf("ticker, batch: %d", len(batch))
			}
			flush()
		}
	}
}