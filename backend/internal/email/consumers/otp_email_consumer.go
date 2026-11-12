package consumers

import (
	"context"
	"encoding/json"

	log "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/mailer"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"go.uber.org/zap"
)

type OtpEmailConsumer struct {
	emailHelper *mailer.EmailHelper
}

func NewOtpEmailConsumer(emailHelper *mailer.EmailHelper) *OtpEmailConsumer {
	return &OtpEmailConsumer{
		emailHelper: emailHelper,
	}
}

func (h *OtpEmailConsumer) Handle(ctx context.Context, body []byte) error {
	var event events.OtpRequestedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Logger.Error("Failed to unmarshal OtpRequestedEvent", zap.Error(err))
		return err
	}

	log.Logger.Info("Sending OTP email", zap.String("email", event.Email))

	if err := h.emailHelper.SendOtpEmail(mailer.OtpEmailData{
		To:      event.Email,
		Otp:     event.Otp,
		Purpose: event.Purpose,
	}); err != nil {
		log.Logger.Error("Failed to send OTP email", zap.Error(err))
		return err
	}

	return nil
}
