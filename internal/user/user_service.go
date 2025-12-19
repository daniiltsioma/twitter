package userservice

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/daniiltsioma/twitter/internal/models"
	"github.com/go-chi/jwtauth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	Register(ctx context.Context, username, password string) (user *models.User, err error)
	Login(ctx context.Context, username, password string) (tokenString string, err error)

	Follow(ctx context.Context, followerId, followedId int64) error
	Unfollow(ctx context.Context, followerId, followedId int64) error
}

type userService struct {
	repo *gorm.DB
	tokenAuth *jwtauth.JWTAuth
}

func NewUserService(repo *gorm.DB, tokenAuth *jwtauth.JWTAuth) *userService {
	return &userService{
		repo: repo, 
		tokenAuth: tokenAuth,
	}
}

func (s *userService) Register(ctx context.Context, username, password string) (*models.User, error) {
	// hash password
	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %v", err)
	}

	user := models.User{
		Username: username,
		PasswordHash: string(pwHash),
	}

	err = gorm.G[models.User](s.repo, gorm.WithResult()).Create(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	return &user, err
}

func (s *userService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := gorm.G[models.User](s.repo).Where("username = ?", username).First(ctx)
	if err != nil {
		log.Printf("user not found: %s", username)
		return "", fmt.Errorf("user not found")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Printf("wrong password for %s", username)
		return "", fmt.Errorf("invalid credentials")
	}

	_, tokenString, err := s.tokenAuth.Encode(map[string]interface{}{"user_id": user.ID})
	if err != nil {
		log.Printf("jwt error: %v", err)
		return "", fmt.Errorf("internal error")
	}

	return tokenString, nil
}

func (s *userService) Follow(ctx context.Context, followerId, followedId int64) error {
	if followerId == followedId {
		return errors.New("userId cannot be the same as targetUserId")
	}

	follow := models.Follow{
		FollowerID: followerId,
		FollowedID: followedId,
	}

	if err := gorm.G[models.Follow](s.repo, gorm.WithResult()).Create(ctx, &follow); err != nil {
		return errors.New("could not follow user")
	}

	return nil
}

func (s *userService) Unfollow(ctx context.Context, followerId, followedId int64) error {
	if followerId == followedId {
		return errors.New("userId cannot be the same as targetUserId")
	}

	n, err := gorm.G[models.Follow](s.repo).Where("follower_id = ? AND followed_id = ?", followerId, followedId).Delete(ctx)
	if err != nil {
		return errors.New("could not unfollow user")
	}
	if n == 0 {
		return errors.New("could not unfollow user")
	}
	return nil
}