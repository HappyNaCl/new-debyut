package models

import "time"

// SearchOutbox records that an event's search document may be out of date.
// Rows are written in the same transaction as the change that caused them, so
// the search index can't miss a committed change; the indexer re-reads the
// event when it processes a row, so a row only names the event, never the
// change itself, and duplicates or reordering are harmless.
type SearchOutbox struct {
	Id      int64  `gorm:"primaryKey;autoIncrement"`
	EventId string `gorm:"type:varchar(20);not null"`
	// The indexer skips the row until then; pushed back after a failed attempt.
	AvailableAt time.Time `gorm:"not null;index"`
	Attempts    int       `gorm:"not null;default:0"`
	LastError   string    `gorm:"type:text"`
	CreatedAt   time.Time
}

func (SearchOutbox) TableName() string {
	return "search_outbox"
}
