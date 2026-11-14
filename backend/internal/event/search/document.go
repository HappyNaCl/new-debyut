// Package search keeps the Meilisearch event index in step with the event
// database (through the search outbox) and runs public event searches.
package search

import (
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/event/models"
	"github.com/microcosm-cc/bluemonday"
)

const IndexName = "events"

// Document is what the index stores per event: the text people search and the
// fields results are filtered and sorted on. Times are Unix seconds, since
// Meilisearch filters compare numbers. Results are rendered from the database,
// not from the document.
type Document struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	OrganizationId string `json:"organizationId"`
	OrganizerName  string `json:"organizerName"`
	CategoryId     string `json:"categoryId"`
	CategorySlug   string `json:"categorySlug"`
	CategoryName   string `json:"categoryName"`
	VenueName      string `json:"venueName"`
	Address        string `json:"address"`
	City           string `json:"city"`
	Description    string `json:"description"`
	StartsAt       int64  `json:"startsAt"`
	// When the event ends; its start when it has no end.
	EndsAt      int64 `json:"endsAt"`
	PublishedAt int64 `json:"publishedAt"`
	// Lowest ticket price in rupiah (0 with a free ticket); null without tickets.
	StartingPrice *int64 `json:"startingPrice"`
	IsFree        bool   `json:"isFree"`
	HasVenue      bool   `json:"hasVenue"`
	HasStreaming  bool   `json:"hasStreaming"`
	Geo           *Geo   `json:"_geo,omitempty"`
}

type Geo struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Settings are applied to the index at startup; changing them makes
// Meilisearch reindex its documents on its own.
var Settings = map[string]any{
	// Ordered by importance: a match in the name ranks above one in the text.
	"searchableAttributes": []string{"name", "organizerName", "categoryName", "venueName", "city", "address", "description"},
	"filterableAttributes": []string{
		"categoryId", "categorySlug", "city", "organizationId",
		"startsAt", "endsAt", "startingPrice", "isFree", "hasVenue", "hasStreaming", "_geo",
	},
	"sortableAttributes":  []string{"startsAt", "publishedAt", "startingPrice", "_geo"},
	"displayedAttributes": []string{"id"},
	"pagination":          map[string]int{"maxTotalHits": 1000},
}

const maxDescriptionRunes = 5000

// Indexable reports whether an event belongs in the index: published and not
// yet ended, the same rule the public listings use.
func Indexable(event *models.Event, now time.Time) bool {
	if event.Status != models.EventStatusPublished || event.StartsAt == nil {
		return false
	}
	return !effectiveEnd(event).Before(now)
}

func effectiveEnd(event *models.Event) time.Time {
	if event.EndsAt != nil {
		return *event.EndsAt
	}
	return *event.StartsAt
}

// NewDocument builds the search document for an indexable event.
func NewDocument(event *models.Event, organizerName string) Document {
	document := Document{
		Id:             event.Id,
		Name:           event.Name,
		OrganizationId: event.OrganizationId,
		OrganizerName:  organizerName,
		VenueName:      event.VenueName,
		Address:        event.Address,
		City:           event.City,
		Description:    PlainText(event.Description, maxDescriptionRunes),
		StartsAt:       event.StartsAt.Unix(),
		EndsAt:         effectiveEnd(event).Unix(),
	}
	if event.PublishedAt != nil {
		document.PublishedAt = event.PublishedAt.Unix()
	}
	if event.CategoryId != nil {
		document.CategoryId = *event.CategoryId
	}
	if event.Category != nil {
		document.CategorySlug = event.Category.Slug
		document.CategoryName = event.Category.Name
	}
	if event.Latitude != nil && event.Longitude != nil {
		document.Geo = &Geo{Lat: *event.Latitude, Lng: *event.Longitude}
	}
	for _, ticket := range event.Tickets {
		price := ticket.Price
		if ticket.Kind == models.TicketFree {
			price = 0
			document.IsFree = true
		}
		if document.StartingPrice == nil || price < *document.StartingPrice {
			document.StartingPrice = &price
		}
		if ticket.Kind == models.TicketStreaming {
			document.HasStreaming = true
		} else {
			document.HasVenue = true
		}
	}
	return document
}

var (
	textPolicy = bluemonday.StrictPolicy()
	whitespace = regexp.MustCompile(`\s+`)
)

// PlainText turns sanitized rich text into searchable text of at most limit
// runes. Tags become spaces so words in adjacent blocks don't run together.
func PlainText(richText string, limit int) string {
	text := html.UnescapeString(textPolicy.Sanitize(strings.ReplaceAll(richText, "<", " <")))
	text = strings.TrimSpace(whitespace.ReplaceAllString(text, " "))
	if runes := []rune(text); len(runes) > limit {
		text = string(runes[:limit])
	}
	return text
}
