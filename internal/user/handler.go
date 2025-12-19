package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
)

type UserHandler struct {
	svc UserService
}

func NewHandler(svc UserService) *UserHandler {
	return &UserHandler{svc}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if in.Username == "" || in.Password == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), in.Username, in.Password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if in.Username == "" || in.Password == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	tokenString, err := h.svc.Login(r.Context(), in.Username, in.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (h *UserHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := int64(claims["user_id"].(float64))

	targetUserId, err := strconv.Atoi(chi.URLParam(r, "targetUserId")); if err != nil {
		http.Error(w, "invalid target user id, must be integer", http.StatusBadRequest)
		return
	}

	if err := h.svc.Follow(r.Context(), userId, int64(targetUserId)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"followerId": userId,
		"followedId": targetUserId,
	})
}

func (h *UserHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := int64(claims["user_id"].(float64))

	targetUserId, err := strconv.Atoi(chi.URLParam(r, "targetUserId")); if err != nil {
		http.Error(w, "invalid target user id, must be integer", http.StatusBadRequest)
		return
	}

	if err := h.svc.Unfollow(r.Context(), userId, int64(targetUserId)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"unfollow": "success"})
}