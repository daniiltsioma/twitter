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

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("twitter.db"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	db.AutoMigrate(&Tweet{})

	r := chi.NewRouter()

	r.Post("/tweet", postTweet)
	r.Get("/tweet/{tweetID}", getTweet)

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