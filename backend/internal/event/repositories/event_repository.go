package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/event/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Tickets", "Category").Create(event).Error; err != nil {
			return err
		}
		return createTickets(tx, event)
	})
}

// UpdateDraft replaces a draft's fields and tickets. It fails with
// ErrEventNotEditable once the event is published.
func (r *EventRepository) UpdateDraft(ctx context.Context, event *models.Event) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.Event
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND organization_id = ?", event.Id, event.OrganizationId).
			First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrEventNotFound
		}
		if err != nil {
			return err
		}
		if current.Status != models.EventStatusDraft {
			return apperrors.ErrEventNotEditable
		}

		event.CreatedBy = current.CreatedBy
		event.CreatedAt = current.CreatedAt
		if err := tx.Omit("Tickets", "Category", "CreatedAt", "CreatedBy").Select("*").Save(event).Error; err != nil {
			return err
		}
		if err := tx.Where("event_id = ?", event.Id).Delete(&models.TicketType{}).Error; err != nil {
			return err
		}
		return createTickets(tx, event)
	})
}

func (r *EventRepository) Get(ctx context.Context, organizationId, id string) (*models.Event, error) {
	var event models.Event
	err := r.db.WithContext(ctx).
		Preload("Tickets", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		Where("id = ? AND organization_id = ?", id, organizationId).
		First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetPublished returns a published event by id; drafts are reported as not found.
func (r *EventRepository) GetPublished(ctx context.Context, id string) (*models.Event, error) {
	var event models.Event
	err := r.db.WithContext(ctx).
		Preload("Tickets", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		Where("id = ? AND status = ?", id, models.EventStatusPublished).
		First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ListLatest returns the most recently published events that haven't ended
// (an event without an end counts as ended once it starts), newest first.
func (r *EventRepository) ListLatest(ctx context.Context, limit int) ([]models.Event, error) {
	var events []models.Event
	err := r.db.WithContext(ctx).
		Preload("Tickets", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		Where("status = ? AND COALESCE(ends_at, starts_at) >= ?", models.EventStatusPublished, time.Now().UTC()).
		Order("published_at DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

type EventFilter string

const (
	EventFilterDraft  EventFilter = "DRAFT"
	EventFilterActive EventFilter = "ACTIVE"
	EventFilterPast   EventFilter = "PAST"
)

// List returns an organization's events for one tab of the dashboard. An
// event counts as past once its end (or start, when it has no end) is over.
func (r *EventRepository) List(ctx context.Context, organizationId string, filter EventFilter) ([]models.Event, error) {
	now := time.Now().UTC()
	query := r.db.WithContext(ctx).
		Preload("Tickets", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		Where("organization_id = ?", organizationId)

	switch filter {
	case EventFilterDraft:
		query = query.Where("status = ?", models.EventStatusDraft).Order("updated_at DESC")
	case EventFilterActive:
		query = query.Where("status = ? AND COALESCE(ends_at, starts_at) >= ?", models.EventStatusPublished, now).Order("starts_at ASC")
	case EventFilterPast:
		query = query.Where("status = ? AND COALESCE(ends_at, starts_at) < ?", models.EventStatusPublished, now).Order("starts_at DESC")
	default:
		return nil, apperrors.ErrInvalidEventFilter
	}

	var events []models.Event
	err := query.Find(&events).Error
	return events, err
}

func createTickets(tx *gorm.DB, event *models.Event) error {
	if len(event.Tickets) == 0 {
		return nil
	}
	for index := range event.Tickets {
		event.Tickets[index].EventId = event.Id
		event.Tickets[index].Position = index
	}
	return tx.Create(&event.Tickets).Error
}
