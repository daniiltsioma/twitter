package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/daniiltsioma/twitter/internal/handler"
	"github.com/daniiltsioma/twitter/internal/models"
	tweetservice "github.com/daniiltsioma/twitter/internal/tweet"
	userservice "github.com/daniiltsioma/twitter/internal/user"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
}

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("twitter.db?_foreign_keys=on"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	db.AutoMigrate(&models.Tweet{}, &models.User{}, &models.Follow{})

	tweetService := tweetservice.NewTweetService(db)
	userService := userservice.NewUserService(db, tokenAuth)

	tweetHandler := handler.NewTweetHandler(tweetService)
	userHandler := handler.NewUserHandler(userService)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator)
		
			r.Post("/tweet", tweetHandler.PostTweet)

			r.Post("/follow/{targetUserId}", userHandler.FollowUser)
			r.Delete("/follow/{targetUserId}", userHandler.UnfollowUser)

			r.Get("/timeline", tweetHandler.GetTimeline)
		})
		
		r.Group(func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)
		})
	})

	fmt.Printf("server listening on port 8080\n")
	if err = http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("%v\n", err)
	}
}
