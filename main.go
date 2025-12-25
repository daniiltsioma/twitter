package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/daniiltsioma/twitter/internal/auth"
	"github.com/daniiltsioma/twitter/internal/timeline"
	"github.com/daniiltsioma/twitter/internal/tweet"
	"github.com/daniiltsioma/twitter/internal/user"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)
}

func main() {
	var err error

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=postgres port=5432 user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	db.AutoMigrate(&tweet.Tweet{}, &user.User{}, &user.Follow{}, &auth.Credentials{})

	// app context
	ctx := context.Background()

	authRepo := auth.NewRepo(db)
	userRepo := user.NewRepo(db)
	tweetRepo := tweet.NewRepo(db)

	tweetService := tweet.NewService(ctx, tweetRepo)
	userService := user.NewService(userRepo)
	authService := auth.NewService(authRepo, userService, tokenAuth)
	timelineService := timeline.NewService(tweetService, userService)

	tweetHandler := tweet.NewHandler(ctx, tweetService)
	userHandler := user.NewHandler(userService)
	authHandler := auth.NewHandler(authService)
	timelineHandler := timeline.NewHandler(timelineService)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(auth.Authenticator)
		
			r.Post("/tweet", tweetHandler.PostTweet)
			
			r.Post("/follow/{targetUserId}", userHandler.FollowUser)
			r.Delete("/follow/{targetUserId}", userHandler.UnfollowUser)
			
			r.Get("/timeline", timelineHandler.GetTweets)
		})
		
		r.Group(func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)

			r.Get("/tweet/{tweetID}", tweetHandler.GetTweet)
		})
	})

	fmt.Printf("server listening on port 8080\n")
	if err = http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("%v\n", err)
	}
}
