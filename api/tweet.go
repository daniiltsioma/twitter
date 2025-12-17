package tweetapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/daniiltsioma/twitter/models"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)


type TweetAPI struct {
	db *gorm.DB
	tokenAuth *jwtauth.JWTAuth
}

func NewTweetAPI(db *gorm.DB, tokenAuth *jwtauth.JWTAuth) *TweetAPI {
	return &TweetAPI{
		db: db,
		tokenAuth: tokenAuth,
	}
}

func (a *TweetAPI) Register(w http.ResponseWriter, r *http.Request) {
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

	// hash password
	pwHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Username: in.Username,
		PasswordHash: string(pwHash),
	}

	err = gorm.G[models.User](a.db, gorm.WithResult()).Create(r.Context(), &user)
	if err != nil {
		log.Printf("error creating user: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (a *TweetAPI) Login(w http.ResponseWriter, r *http.Request) {
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

	user, err := gorm.G[models.User](a.db).Where("username = ?", in.Username).First(r.Context())
	if err != nil {
		log.Printf("user not found: %s", in.Username)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		log.Printf("wrong password for %s", in.Username)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	_, tokenString, err := a.tokenAuth.Encode(map[string]interface{}{"user_id": user.ID})
	if err != nil {
		log.Printf("jwt error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (a *TweetAPI) PostTweet(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := int64(claims["user_id"].(float64))

	var tweet models.Tweet

	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		http.Error(w, "invalid JSON: " + err.Error(), http.StatusBadRequest)
		return
	}

	tweet.UserID = userId

	if tweet.Text == "" {
		http.Error(w, "tweet.Text cannot be empty", http.StatusBadRequest)
		return
	}

	if err := gorm.G[models.Tweet](a.db, gorm.WithResult()).Create(r.Context(), &tweet); err != nil {
		http.Error(w, "could not save tweet: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tweet)
}

func (a *TweetAPI) GetTweet(w http.ResponseWriter, r *http.Request) {
	tweetID := chi.URLParam(r, "tweetID")
	
	tweet, err := gorm.G[models.Tweet](a.db).Where("ID = ?", tweetID).First(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweet)
}

func (a *TweetAPI) GetUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := gorm.G[models.User](a.db).Where("username = ?", username).First(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (a *TweetAPI) FollowUser(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := int64(claims["user_id"].(float64))

	var follow models.Follow

	if err := json.NewDecoder(r.Body).Decode(&follow); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	follow.FollowerID = userId

	if follow.FollowedID == 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	if follow.FollowerID == follow.FollowedID {
		http.Error(w, "userId cannot be the same as targetUserId", http.StatusBadRequest)
		return
	}

	if err := gorm.G[models.Follow](a.db, gorm.WithResult()).Create(r.Context(), &follow); err != nil {
		http.Error(w, "could not follow user: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(follow)
}

func (a *TweetAPI) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := int64(claims["user_id"].(float64))

	var unfollow models.Follow

	if err := json.NewDecoder(r.Body).Decode(&unfollow); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	unfollow.FollowerID = userId

	if unfollow.FollowedID == 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	if unfollow.FollowerID == unfollow.FollowedID {
		http.Error(w, "userId cannot be the same as targetUserId", http.StatusBadRequest)
		return
	}

	n, err := gorm.G[models.Follow](a.db).Where("follower_id = ? AND followed_id = ?", unfollow.FollowerID, unfollow.FollowedID).Delete(r.Context())
	if err != nil {
		http.Error(w, "could not unfollow user: " + err.Error(), http.StatusInternalServerError)
		return
	}
	if n == 0 {
		http.Error(w, "no follow record found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"unfollow": "success"})
}

func (a *TweetAPI) GetUserTimeline(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userID := int64(claims["user_id"].(float64))

	var tweets []models.Tweet

	a.db.Joins("JOIN follows on follows.followed_id = tweets.user_id").Where("follows.follower_id = ?", userID).Find(&tweets)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweets)
}
