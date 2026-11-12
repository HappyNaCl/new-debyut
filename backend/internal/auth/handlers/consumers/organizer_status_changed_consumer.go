package consumers

import (
	"context"
	"encoding/json"

	"github.com/HappyNaCl/new-debyut/internal/auth/services"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"go.uber.org/zap"
)

type OrganizerStatusChangedConsumer struct {
	authService *services.AuthService
}

func NewOrganizerStatusChangedConsumer(authService *services.AuthService) *OrganizerStatusChangedConsumer {
	return &OrganizerStatusChangedConsumer{
		authService: authService,
	}
}

func (h *OrganizerStatusChangedConsumer) Handle(ctx context.Context, body []byte) error {
	var event events.OrganizerStatusChangedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		l.Logger.Error("Failed to unmarshal OrganizerStatusChangedEvent", zap.Error(err))
		return nil
	}

	if err := h.authService.SetOrganizerStatus(ctx, event.UserId, event.IsOrganizer); err != nil {
		return err
	}

	l.Logger.Info("Processed OrganizerStatusChangedEvent", zap.String("userId", event.UserId), zap.Bool("isOrganizer", event.IsOrganizer))
	return nil
}
