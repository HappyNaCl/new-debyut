package consumers

import (
	"context"
	"encoding/json"
	"os"

	log "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/mailer"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"go.uber.org/zap"
)

type InvitationEmailConsumer struct {
	emailHelper *mailer.EmailHelper
}

func NewInvitationEmailConsumer(emailHelper *mailer.EmailHelper) *InvitationEmailConsumer {
	return &InvitationEmailConsumer{
		emailHelper: emailHelper,
	}
}

func (h *InvitationEmailConsumer) Handle(ctx context.Context, body []byte) error {
	var event events.OrganizationInvitationCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Logger.Error("Failed to unmarshal OrganizationInvitationCreatedEvent", zap.Error(err))
		return nil
	}

	log.Logger.Info("Sending invitation email", zap.String("email", event.Email), zap.String("invitationId", event.InvitationId))

	if err := h.emailHelper.SendInvitationEmail(mailer.InvitationEmailData{
		To:               event.Email,
		OrganizationName: event.OrganizationName,
		InviterName:      event.InviterName,
		Role:             event.Role,
		ExpiresAt:        event.ExpiresAt.Format("2 January 2006"),
		InvitationUrl:    os.Getenv("FRONTEND_URL") + "/#/invitations",
	}); err != nil {
		log.Logger.Error("Failed to send invitation email", zap.Error(err))
		return err
	}

	return nil
}
