package consumers

import (
	"context"
	"encoding/json"

	log "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/mailer"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"go.uber.org/zap"
)

type WelcomeUserEmailConsumer struct {
	emailHelper *mailer.EmailHelper
}

func NewWelcomeUserEmailConsumer(emailHelper *mailer.EmailHelper) *WelcomeUserEmailConsumer {
	return &WelcomeUserEmailConsumer{
		emailHelper: emailHelper,
	}
}

func (h *WelcomeUserEmailConsumer) Handle(ctx context.Context, body []byte) error {
	var event events.UserRegisteredEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Logger.Error("Failed to unmarshal UserRegisteredEvent", zap.Error(err))
		return err
	}

	log.Logger.Info("Sending welcome email", zap.String("email", event.Email))

	if err := h.emailHelper.SendWelcomeEmail(mailer.WelcomeEmailData{
		To:        event.Email,
		FirstName: event.FirstName,
	}); err != nil {
		log.Logger.Error("Failed to send email", zap.Error(err))
		return err
	}

	return nil
}
