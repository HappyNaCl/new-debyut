package consumers

import (
	"context"
	"encoding/json"

	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"github.com/HappyNaCl/new-debyut/internal/users/services"
	"go.uber.org/zap"
)

type UserRegisteredConsumer struct {
	userService *services.UserService
}

func NewUserRegisteredConsumer(userService *services.UserService) *UserRegisteredConsumer {
	return &UserRegisteredConsumer{
		userService: userService,
	}
}

func (h *UserRegisteredConsumer) Handle(ctx context.Context, body []byte) error {
	var event events.UserRegisteredEvent
	if err := json.Unmarshal(body, &event); err != nil {
		l.Logger.Error("Failed to unmarshal UserRegisteredEvent", zap.Error(err))
		return err
	}

	newUser, err := h.userService.CreateUser(ctx, event)
	if err != nil {
		return err
	}

	l.Logger.Info("Processed UserRegisteredEvent", zap.String("userId", newUser.Id))
	return nil
}
