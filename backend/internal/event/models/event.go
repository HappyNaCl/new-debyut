package models

import "time"

type EventStatus string

const (
	EventStatusDraft     EventStatus = "DRAFT"
	EventStatusPublished EventStatus = "PUBLISHED"
)

type ServiceFeeBearer string

const (
	ServiceFeeOrganizer ServiceFeeBearer = "ORGANIZER"
	ServiceFeeBuyer     ServiceFeeBearer = "BUYER"
)

// Event is owned by an organization in the users service; OrganizationId and
// CreatedBy are not foreign keys because that data lives in another database.
// Drafts may leave most fields empty; publishing requires a complete event.
// The venue fields are required only when some ticket grants venue entry:
// an event sold purely as STREAMING tickets has no venue.
type Event struct {
	Id               string      `gorm:"type:varchar(20);primaryKey"`
	OrganizationId   string      `gorm:"type:varchar(20);not null;index"`
	CreatedBy        string      `gorm:"type:varchar(20);not null"`
	Status           EventStatus `gorm:"type:varchar(10);not null;default:DRAFT;index"`
	Name             string      `gorm:"type:varchar(150);not null"`
	CategoryId       *string     `gorm:"type:varchar(20);index"`
	Category         *Category   `gorm:"foreignKey:CategoryId;constraint:OnDelete:RESTRICT"`
	BannerUrl        string      `gorm:"type:varchar(512)"`
	CoverUrl         string      `gorm:"type:varchar(512)"`
	GalleryUrls      []string    `gorm:"type:jsonb;serializer:json"`
	StartsAt         *time.Time  `gorm:"index"`
	EndsAt           *time.Time  `gorm:"index"`
	VenueName        string      `gorm:"type:varchar(150)"`
	Address          string      `gorm:"type:varchar(500)"`
	City             string      `gorm:"type:varchar(100)"`
	Latitude         *float64    // venue pin; set together with Longitude
	Longitude        *float64
	Description      string           `gorm:"type:text"` // sanitized HTML
	Terms            string           `gorm:"type:text"` // sanitized HTML
	ContactName      string           `gorm:"type:varchar(100)"`
	ContactEmail     string           `gorm:"type:varchar(255)"`
	ContactPhone     string           `gorm:"type:varchar(20)"`
	ServiceFeeBearer ServiceFeeBearer `gorm:"type:varchar(10);not null;default:ORGANIZER"`
	Tickets          []TicketType     `gorm:"foreignKey:EventId;constraint:OnDelete:CASCADE"`
	PublishedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type TicketKind string

const (
	TicketPaid TicketKind = "PAID"
	TicketFree TicketKind = "FREE"
	// TicketStreaming grants online viewing instead of venue entry. Always
	// paid; playback will be hosted by debyut (not an organizer link).
	TicketStreaming TicketKind = "STREAMING"
)

type TicketType struct {
	Id           string     `gorm:"type:varchar(20);primaryKey"`
	EventId      string     `gorm:"type:varchar(20);not null;index"`
	Position     int        `gorm:"not null;default:0"`
	Kind         TicketKind `gorm:"type:varchar(20);not null"`
	Name         string     `gorm:"type:varchar(100);not null"`
	Description  string     `gorm:"type:varchar(500)"`
	Price        int64      `gorm:"not null;default:0"`
	Quantity     int32      `gorm:"not null"`
	SalesStartAt *time.Time
	SalesEndAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
