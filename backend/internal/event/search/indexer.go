package search

import (
	"context"
	"fmt"
	"strconv"
	"time"

	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/meilisearch"
	"github.com/HappyNaCl/new-debyut/internal/event/models"
	"github.com/HappyNaCl/new-debyut/internal/event/repositories"
	"go.uber.org/zap"
)

const (
	batchSize    = 100
	pollInterval = time.Second
	sweepEvery   = time.Hour
	maxBackoff   = 5 * time.Minute
	writeTimeout = 30 * time.Second
)

// OrganizerNames resolves organization ids to display names. Organizations
// live in the users service; ids it doesn't know are left out.
type OrganizerNames interface {
	Names(ctx context.Context, organizationIds []string) (map[string]string, error)
}

// Indexer relays the search outbox to Meilisearch. Every committed change to
// a searchable event leaves an outbox row, and a row is only deleted after
// Meilisearch has applied the event's current state, so the index catches up
// after any crash or outage. Several instances can run at once.
type Indexer struct {
	lastWarning time.Time

	client     *meilisearch.Client
	outbox     *repositories.SearchOutboxRepository
	events     *repositories.EventRepository
	organizers OrganizerNames
}

func NewIndexer(
	client *meilisearch.Client,
	outbox *repositories.SearchOutboxRepository,
	events *repositories.EventRepository,
	organizers OrganizerNames,
) *Indexer {
	return &Indexer{client: client, outbox: outbox, events: events, organizers: organizers}
}

// Setup creates the index if needed and applies its settings.
func (i *Indexer) Setup(ctx context.Context) error {
	if _, err := i.client.EnsureIndex(ctx, IndexName, "id", Settings); err != nil {
		return fmt.Errorf("set up search index: %w", err)
	}
	return nil
}

// fillIfEmpty queues every published event when the index has no documents
// (new, wiped, or never filled). Checked at start and on every sweep, so it
// also recovers when the database wasn't migrated yet at startup.
func (i *Indexer) fillIfEmpty(ctx context.Context) {
	count, err := i.client.DocumentCount(ctx, IndexName)
	if err != nil {
		l.Logger.Warn("Search index: failed to count documents", zap.Error(err))
		return
	}
	if count > 0 {
		return
	}
	if err := i.outbox.EnqueueAll(ctx); err != nil {
		l.Logger.Warn("Search index: failed to queue events for indexing", zap.Error(err))
		return
	}
	l.Logger.Info("Search index is empty; queued all published events")
}

// Run relays the outbox until ctx is cancelled. Every hour it removes ended
// events from the index (ending isn't a write, so it leaves no outbox row)
// and refills the index if it's empty.
func (i *Indexer) Run(ctx context.Context) {
	poll := time.NewTicker(pollInterval)
	defer poll.Stop()
	sweep := time.NewTicker(sweepEvery)
	defer sweep.Stop()

	i.fillIfEmpty(ctx)
	i.sweepEnded(ctx)
	for {
		i.drain(ctx)
		select {
		case <-ctx.Done():
			return
		case <-poll.C:
		case <-sweep.C:
			i.sweepEnded(ctx)
			i.fillIfEmpty(ctx)
		}
	}
}

// drain processes batches until the outbox has nothing due or a batch fails.
func (i *Indexer) drain(ctx context.Context) {
	for ctx.Err() == nil {
		claimed, processErr, err := i.outbox.Claim(ctx, batchSize, backoff, func(rows []models.SearchOutbox) error {
			return i.process(ctx, rows)
		})
		if err != nil {
			i.warn("Search outbox: database error", zap.Error(err))
			return
		}
		if processErr != nil {
			i.warn("Search outbox: batch failed, will retry", zap.Int("rows", claimed), zap.Error(processErr))
			return
		}
		if claimed < batchSize {
			return
		}
	}
}

// process writes the current state of every event named in rows: upserted
// when it's searchable, deleted otherwise.
func (i *Indexer) process(ctx context.Context, rows []models.SearchOutbox) error {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()

	seen := make(map[string]bool, len(rows))
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if !seen[row.EventId] {
			seen[row.EventId] = true
			ids = append(ids, row.EventId)
		}
	}

	events, err := i.events.ListForIndex(ctx, ids)
	if err != nil {
		return fmt.Errorf("load events: %w", err)
	}

	now := time.Now().UTC()
	indexable := make([]*models.Event, 0, len(events))
	keep := make(map[string]bool, len(events))
	organizationIds := make([]string, 0, len(events))
	for index := range events {
		event := &events[index]
		if Indexable(event, now) {
			indexable = append(indexable, event)
			keep[event.Id] = true
			organizationIds = append(organizationIds, event.OrganizationId)
		}
	}

	if len(indexable) > 0 {
		names, err := i.organizers.Names(ctx, organizationIds)
		if err != nil {
			return fmt.Errorf("look up organizers: %w", err)
		}
		documents := make([]Document, 0, len(indexable))
		for _, event := range indexable {
			documents = append(documents, NewDocument(event, names[event.OrganizationId]))
		}
		if err := i.client.UpsertDocuments(ctx, IndexName, documents); err != nil {
			return fmt.Errorf("upsert documents: %w", err)
		}
	}

	removed := make([]string, 0, len(ids))
	for _, id := range ids {
		if !keep[id] {
			removed = append(removed, id)
		}
	}
	if len(removed) > 0 {
		if err := i.client.DeleteDocuments(ctx, IndexName, removed); err != nil {
			return fmt.Errorf("delete documents: %w", err)
		}
	}
	return nil
}

// sweepEnded drops events that have ended since they were indexed. Searches
// filter these out anyway; this just keeps the index small.
func (i *Indexer) sweepEnded(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	filter := "endsAt < " + strconv.FormatInt(time.Now().Unix(), 10)
	if err := i.client.DeleteByFilter(ctx, IndexName, filter); err != nil && ctx.Err() == nil {
		l.Logger.Warn("Search index: failed to remove ended events", zap.Error(err))
	}
}

// warn logs at most once a minute, since the relay polls every second and a
// lasting problem (e.g. an unmigrated database) would otherwise flood the log.
func (i *Indexer) warn(message string, fields ...zap.Field) {
	if time.Since(i.lastWarning) < time.Minute {
		return
	}
	i.lastWarning = time.Now()
	l.Logger.Warn(message, fields...)
}

// backoff doubles from 2s per failed attempt, up to five minutes.
func backoff(attempts int) time.Duration {
	if attempts > 8 {
		return maxBackoff
	}
	return min(time.Second<<attempts, maxBackoff)
}
