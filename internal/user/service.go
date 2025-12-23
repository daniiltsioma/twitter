package user

import (
	"context"
	"errors"
	"fmt"
	"log"
)

type UserService interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)

	Follow(ctx context.Context, followerId, followedId int64) error
	Unfollow(ctx context.Context, followerId, followedId int64) error

	GetFollows(ctx context.Context, userId int64) ([]Follow, error)
}

type userService struct {
	repo UserRepo
}

func NewService(repo UserRepo) *userService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *User) (*User, error) {
	err := s.repo.InsertUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	return user, err
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		log.Printf("user not found: %s", username)
		return nil, fmt.Errorf("user not found")
	}

	return &user, nil
}

func (s *userService) Follow(ctx context.Context, followerId, followedId int64) error {
	if followerId == followedId {
		return errors.New("userId cannot be the same as targetUserId")
	}

	if err := s.repo.InsertFollow(ctx, followerId, followedId); err != nil {
		return errors.New("could not follow user")
	}

	return nil
}

func (s *userService) Unfollow(ctx context.Context, followerId, followedId int64) error {
	if followerId == followedId {
		return errors.New("userId cannot be the same as targetUserId")
	}

	return s.repo.DeleteFollow(ctx, followerId, followedId)
}

func (s *userService) GetFollows(ctx context.Context, userId int64) ([]Follow, error) {
	follows, err := s.repo.GetFollows(ctx, userId)
	if err != nil {
		log.Printf("repo error: %v", err)
		return nil, err
	}

	return follows, err
}