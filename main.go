package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
}

type Tweet struct {
	ID int64 `gorm:"primaryKey"`
	UserID int64 `json:"userId"`
	User User `gorm:"foreignKey:UserID"`
	Text string `json:"text"`
}

type User struct {
	ID int64 `gorm:"primaryKey"`
	Username string `json:"username" gorm:"uniqueIndex"`
	PasswordHash string `json:"-"`
} 

type Follow struct {
	ID int64 `gorm:"primaryKey"`
	FollowerID int64 `json:"followerId"`
	FollowedID int64 `json:"followedId"`
	Follower User `gorm:"foreignKey:FollowerID"`
	Followed User `gorm:"foreignKey:FollowedID"`
}

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("twitter.db?_foreign_keys=on"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	db.AutoMigrate(&Tweet{}, &User{}, &Follow{})

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)
	
		r.Post("/tweet", postTweet)
	})

	r.Group(func(r chi.Router) {
		r.Post("/register", register)
		r.Post("/login", login)
	
		r.Get("/tweet/{tweetID}", getTweet)
	
		r.Get("/user/{username}", getUser)
	
		r.Post("/follow", followUser)
		r.Delete("/follow", unfollowUser)
	
		r.Get("/timeline/{userID}", getUserTimeline)
	})


	fmt.Printf("server listening on port 8080\n")
	http.ListenAndServe(":8080", r)
}

func register(w http.ResponseWriter, r *http.Request) {
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

	user := User{
		Username: in.Username,
		PasswordHash: string(pwHash),
	}

	err = gorm.G[User](db, gorm.WithResult()).Create(r.Context(), &user)
	if err != nil {
		log.Printf("error creating user: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func login(w http.ResponseWriter, r *http.Request) {
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

	user, err := gorm.G[User](db).Where("username = ?", in.Username).First(r.Context())
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

	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{"user_id": user.ID})
	if err != nil {
		log.Printf("jwt error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func postTweet(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Printf("jwt error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var tweet Tweet

	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		http.Error(w, "invalid JSON: " + err.Error(), http.StatusBadRequest)
		return
	}

	tweet.UserID = int64(claims["user_id"].(float64))

	if tweet.Text == "" {
		http.Error(w, "tweet.Text cannot be empty", http.StatusBadRequest)
		return
	}

	if err := gorm.G[Tweet](db, gorm.WithResult()).Create(r.Context(), &tweet); err != nil {
		http.Error(w, "could not save tweet: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tweet)
}

func getTweet(w http.ResponseWriter, r *http.Request) {
	tweetID := chi.URLParam(r, "tweetID")
	
	tweet, err := gorm.G[Tweet](db).Where("ID = ?", tweetID).First(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweet)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := gorm.G[User](db).Where("username = ?", username).First(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func followUser(w http.ResponseWriter, r *http.Request) {
	var follow Follow

	if err := json.NewDecoder(r.Body).Decode(&follow); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if follow.FollowerID == 0 || follow.FollowedID == 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	if follow.FollowerID == follow.FollowedID {
		http.Error(w, "userId cannot be the same as targetUserId", http.StatusBadRequest)
		return
	}

	if err := gorm.G[Follow](db, gorm.WithResult()).Create(r.Context(), &follow); err != nil {
		http.Error(w, "could not follow user: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(follow)
}

func unfollowUser(w http.ResponseWriter, r *http.Request) {
	var unfollow Follow
	if err := json.NewDecoder(r.Body).Decode(&unfollow); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if unfollow.FollowerID == 0 || unfollow.FollowedID == 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	if unfollow.FollowerID == unfollow.FollowedID {
		http.Error(w, "userId cannot be the same as targetUserId", http.StatusBadRequest)
		return
	}

	n, err := gorm.G[Follow](db).Where("follower_id = ? AND followed_id = ?", unfollow.FollowerID, unfollow.FollowedID).Delete(r.Context())
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

func getUserTimeline(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	var tweets []Tweet

	db.Joins("JOIN follows on follows.followed_id = tweets.user_id").Where("follows.follower_id = ?", userID).Find(&tweets)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweets)
}