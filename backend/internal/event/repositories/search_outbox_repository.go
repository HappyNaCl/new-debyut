package repositories

import (
	"context"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/event/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// enqueueEvent marks one event for reindexing. Call it inside the transaction
// that changes the event.
func enqueueEvent(tx *gorm.DB, eventId string) error {
	now := time.Now().UTC()
	return tx.Create(&models.SearchOutbox{EventId: eventId, AvailableAt: now, CreatedAt: now}).Error
}

// enqueuePublished marks every published event matching the condition for
// reindexing, e.g. after something every one of them shows (a category or
// organizer name) changed.
func enqueuePublished(tx *gorm.DB, condition string, args ...any) error {
	query := "INSERT INTO search_outbox (event_id, available_at, attempts, last_error, created_at) " +
		"SELECT id, now(), 0, '', now() FROM events WHERE status = ?"
	values := []any{models.EventStatusPublished}
	if condition != "" {
		query += " AND (" + condition + ")"
		values = append(values, args...)
	}
	return tx.Exec(query, values...).Error
}

type SearchOutboxRepository struct {
	db *gorm.DB
}

func NewSearchOutboxRepository(db *gorm.DB) *SearchOutboxRepository {
	return &SearchOutboxRepository{db: db}
}

// EnqueueOrganization reindexes an organization's published events.
func (r *SearchOutboxRepository) EnqueueOrganization(ctx context.Context, organizationId string) error {
	return enqueuePublished(r.db.WithContext(ctx), "organization_id = ?", organizationId)
}

// EnqueueAll reindexes every published event (to fill or rebuild the index).
func (r *SearchOutboxRepository) EnqueueAll(ctx context.Context) error {
	return enqueuePublished(r.db.WithContext(ctx), "")
}

// Claim locks up to limit due rows for the length of the transaction and
// passes them to process; rows other workers hold are skipped. When process
// succeeds the rows are deleted; when it fails they're kept, with the error
// recorded, and retried after backoff. It returns how many rows were claimed,
// process's error, and any database error.
func (r *SearchOutboxRepository) Claim(
	ctx context.Context,
	limit int,
	backoff func(attempts int) time.Duration,
	process func(rows []models.SearchOutbox) error,
) (claimed int, processErr error, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []models.SearchOutbox
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("available_at <= ?", time.Now().UTC()).
			Order("id ASC").
			Limit(limit).
			Find(&rows).Error
		if err != nil || len(rows) == 0 {
			return err
		}
		claimed = len(rows)

		ids := make([]int64, len(rows))
		for index, row := range rows {
			ids[index] = row.Id
		}

		if processErr = process(rows); processErr != nil {
			// One failure postpones the whole batch; the next attempt re-reads
			// every event, so retrying the ones that would have worked is fine.
			attempts := 1
			for _, row := range rows {
				attempts = max(attempts, row.Attempts+1)
			}
			message := processErr.Error()
			if len(message) > 1000 {
				message = message[:1000]
			}
			return tx.Model(&models.SearchOutbox{}).Where("id IN ?", ids).Updates(map[string]any{
				"attempts":     gorm.Expr("attempts + 1"),
				"available_at": time.Now().UTC().Add(backoff(attempts)),
				"last_error":   message,
			}).Error
		}
		return tx.Where("id IN ?", ids).Delete(&models.SearchOutbox{}).Error
	})
	return claimed, processErr, err
}
