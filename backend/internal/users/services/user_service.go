package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/storage"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/HappyNaCl/new-debyut/internal/users/repositories"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const minimumAge = 18

type ProfileUpdate struct {
	FirstName           string
	Surname             string
	DateOfBirth         time.Time
	PhoneNumber         string
	Gender              models.Gender
	SubscribeNewsletter bool
}

type UserService struct {
	userRepository *repositories.UserRepository
	redis          *redis.Client
	storage        *storage.Client
}

func (s *UserService) getMeCacheKey(id string) string {
	return "me:" + id
}

func NewUserService(userRepo *repositories.UserRepository, redis *redis.Client, storage *storage.Client) *UserService {
	return &UserService{
		userRepository: userRepo,
		redis:          redis,
		storage:        storage,
	}
}

func (s *UserService) CreateUser(ctx context.Context, event events.UserRegisteredEvent) (*models.User, error) {
	newUser := &models.User{
		Id:                  event.UserId,
		FirstName:           event.FirstName,
		Surname:             event.Surname,
		DateOfBirth:         event.DateOfBirth,
		PhoneNumber:         event.PhoneNumber,
		Gender:              models.Gender(event.Gender),
		SubscribeNewsletter: event.SubscribeNewsletter,
		TermsAcceptedAt:     event.TermsAcceptedAt,
	}

	if _, err := s.userRepository.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *UserService) GetMe(ctx context.Context, id string) (*models.User, error) {
	key := s.getMeCacheKey(id)

	cached, err := s.redis.Get(ctx, key).Bytes()
	if err == nil {
		var user models.User
		if err := json.Unmarshal(cached, &user); err == nil {
			return &user, nil
		}
	}

	user, err := s.userRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(user); err == nil {
		if err := s.redis.Set(ctx, key, data, 15*time.Minute).Err(); err != nil {
			l.Logger.Warn("failed to cache user profile", zap.String("userId", id), zap.Error(err))
		}
	}

	return user, nil
}

func (s *UserService) UpdateMe(ctx context.Context, id string, update ProfileUpdate) (*models.User, error) {
	if update.DateOfBirth.AddDate(minimumAge, 0, 0).After(time.Now()) {
		return nil, apperrors.ErrUnderAge
	}

	user, err := s.userRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	user.FirstName = update.FirstName
	user.Surname = update.Surname
	user.DateOfBirth = update.DateOfBirth
	user.PhoneNumber = update.PhoneNumber
	user.Gender = update.Gender
	user.SubscribeNewsletter = update.SubscribeNewsletter

	if err := s.save(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, id string, content []byte) (*models.User, error) {
	user, err := s.userRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	url, err := uploadAvatar(ctx, s.storage, "users", id, content)
	if err != nil {
		return nil, err
	}

	previousUrl := user.AvatarUrl
	user.AvatarUrl = url
	if err := s.save(ctx, user); err != nil {
		deleteAvatar(ctx, s.storage, url)
		return nil, err
	}

	deleteAvatar(ctx, s.storage, previousUrl)
	return user, nil
}

func (s *UserService) save(ctx context.Context, user *models.User) error {
	if err := s.userRepository.Update(ctx, user); err != nil {
		l.Logger.Error("failed to update user profile", zap.String("userId", user.Id), zap.Error(err))
		return apperrors.ErrUserFailedUpdate
	}

	if err := s.redis.Del(ctx, s.getMeCacheKey(user.Id)).Err(); err != nil {
		l.Logger.Warn("failed to evict cached user profile", zap.String("userId", user.Id), zap.Error(err))
	}
	return nil
}
