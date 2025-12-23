package timeline

import (
	"encoding/json"
	"net/http"

	"github.com/daniiltsioma/twitter/internal/auth"
)

type TimelineHandler struct {
	svc TimelineService
}

func NewHandler(svc TimelineService) *TimelineHandler {
	return &TimelineHandler{svc: svc}
}

func (h *TimelineHandler) GetTweets(w http.ResponseWriter, r *http.Request) {
	userId, ok := auth.UserIDFromContext(r.Context())		
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tweets, err := h.svc.GetTweets(r.Context(), userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweets)
}