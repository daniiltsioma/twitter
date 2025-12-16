package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Tweet struct {
	ID int64 	`json:"id"`
	Text string `json:"text"`
}

type User struct {
	ID int64 `json:"id"`
	Username string `json:"username" gorm:"uniqueIndex"`
} 

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("twitter.db"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	db.AutoMigrate(&Tweet{}, &User{})

	r := chi.NewRouter()

	r.Post("/tweet", postTweet)
	r.Get("/tweet/{tweetID}", getTweet)

	r.Post("/user", createUser)
	r.Get("/user/{username}", getUser)

	fmt.Printf("server listening on port 8080\n")
	http.ListenAndServe(":8080", r)
}

func postTweet(w http.ResponseWriter, r *http.Request) {
	var tweet Tweet
	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		http.Error(w, "invalid JSON: " + err.Error(), http.StatusBadRequest)
		return
	}

	if tweet.Text == "" {
		http.Error(w, "tweet text cannot be empty", http.StatusBadRequest)
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

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid JSON" + err.Error(), http.StatusBadRequest)
		return
	}

	if user.Username == "" {
		http.Error(w, "username cannot be empty", http.StatusBadRequest)
		return
	}

	if err := gorm.G[User](db, gorm.WithResult()).Create(r.Context(), &user); err != nil {
		http.Error(w, "could not create user:" + err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
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