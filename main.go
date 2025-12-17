package main

import (
	"fmt"
	"log"
	"net/http"

	tweetapi "github.com/daniiltsioma/twitter/api"
	"github.com/daniiltsioma/twitter/models"
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

	api := tweetapi.NewTweetAPI(db, tokenAuth)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator)
		
			r.Post("/tweet", api.PostTweet)
			r.Post("/follow", api.FollowUser)
			r.Delete("/follow", api.UnfollowUser)
			r.Get("/timeline", api.GetUserTimeline)
		})
		
		r.Group(func(r chi.Router) {
			r.Post("/register", api.Register)
			r.Post("/login", api.Login)
			
			r.Get("/tweet/{tweetID}", api.GetTweet)
			r.Get("/user/{username}", api.GetUser)
		})
	})

	r.Route("/", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		
		r.Get("/", handleHomePage)
	})


	fmt.Printf("server listening on port 8080\n")
	http.ListenAndServe(":8080", r)
}

func handleHomePage(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	if userId := claims["user_id"]; userId != nil {
		RenderTimeline(w, r, int64(userId.(float64)))
		return
	} 
	RenderLogin(w, r)	
}

func RenderTimeline(w http.ResponseWriter, r *http.Request, userId int64) {
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "html/timeline.html")
}

func RenderLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "html/login.html")
}