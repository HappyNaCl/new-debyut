package consumers

import (
	"context"
	"encoding/json"

	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/event/repositories"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"go.uber.org/zap"
)

// OrganizationRenamedConsumer queues an organization's published events for
// reindexing when it's renamed, since search documents carry organizer names.
type OrganizationRenamedConsumer struct {
	outbox *repositories.SearchOutboxRepository
}

func NewOrganizationRenamedConsumer(outbox *repositories.SearchOutboxRepository) *OrganizationRenamedConsumer {
	return &OrganizationRenamedConsumer{outbox: outbox}
}

func (c *OrganizationRenamedConsumer) Handle(ctx context.Context, body []byte) error {
	var event events.OrganizationRenamedEvent
	if err := json.Unmarshal(body, &event); err != nil || event.OrganizationId == "" {
		// Malformed messages would be redelivered forever; drop them.
		l.Logger.Warn("Dropping malformed organization renamed message", zap.ByteString("body", body))
		return nil
	}
	return c.outbox.EnqueueOrganization(ctx, event.OrganizationId)
}
