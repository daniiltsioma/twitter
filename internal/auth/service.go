package auth

import (
	"context"
	"fmt"
	"log"

	"github.com/daniiltsioma/twitter/internal/user"
	"github.com/go-chi/jwtauth"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (userId int64, err error)
	Login(ctx context.Context, username, password string) (tokenString string, err error)
}

type authService struct {
	repo AuthRepo
	us user.UserService
	tokenAuth *jwtauth.JWTAuth
}

func NewService(repo AuthRepo, us user.UserService, tokenAuth *jwtauth.JWTAuth) *authService {
	return &authService{
		repo: repo, 
		us: us,
		tokenAuth: tokenAuth,
	}
}

func (s *authService) Register(ctx context.Context, username, password string) (userId int64, err error) {
	// hash password
	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("error hashing password: %v", err)
	}

	user := user.User{
		Username: username,
	}

	// TODO: update CreateUser signature
	_, err = s.us.CreateUser(ctx, &user)
	if err != nil {
		log.Printf("error creating user: %v", err)
		return 0, err
	}
	
	err = s.repo.InsertCredentials(ctx, user.ID, string(pwHash))
	if err != nil {
		log.Printf("error inserting credentials: %v", err)
		return 0, err
	}

	return user.ID, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (tokenString string, err error) {
	user, err := s.us.GetByUsername(ctx, username)
	if err != nil {
		log.Printf("user not found: %s", username)
		return "", fmt.Errorf("user not found")
	}

	pwHash, err := s.repo.GetPasswordHash(ctx, user.ID)
	if err != nil {
		return "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(pwHash), []byte(password)); err != nil {
		log.Printf("wrong password for %s", username)
		return "", fmt.Errorf("invalid credentials")
	}

	_, tokenString, err = s.tokenAuth.Encode(map[string]interface{}{"user_id": user.ID})
	if err != nil {
		log.Printf("jwt error: %v", err)
		return "", fmt.Errorf("internal error")
	}

	return tokenString, nil
}