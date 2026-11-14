package search

import (
	"testing"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/event/models"
)

func TestIndexable(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	past, future := now.Add(-time.Hour), now.Add(time.Hour)

	cases := []struct {
		name  string
		event models.Event
		want  bool
	}{
		{"draft", models.Event{Status: models.EventStatusDraft, StartsAt: &future}, false},
		{"upcoming", models.Event{Status: models.EventStatusPublished, StartsAt: &future}, true},
		{"running", models.Event{Status: models.EventStatusPublished, StartsAt: &past, EndsAt: &future}, true},
		{"ended", models.Event{Status: models.EventStatusPublished, StartsAt: &past, EndsAt: &past}, false},
		{"started without end", models.Event{Status: models.EventStatusPublished, StartsAt: &past}, false},
		{"no start", models.Event{Status: models.EventStatusPublished}, false},
	}
	for _, tc := range cases {
		if got := Indexable(&tc.event, now); got != tc.want {
			t.Errorf("%s: Indexable = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestNewDocument(t *testing.T) {
	starts := time.Date(2026, 12, 1, 19, 0, 0, 0, time.UTC)
	lat, lng := -6.2, 106.8
	category := "c1"
	event := models.Event{
		Id:             "e1",
		Name:           "Jazz Night",
		OrganizationId: "o1",
		Status:         models.EventStatusPublished,
		CategoryId:     &category,
		Category:       &models.Category{Id: category, Slug: "music", Name: "Music"},
		StartsAt:       &starts,
		City:           "Jakarta",
		Latitude:       &lat,
		Longitude:      &lng,
		Description:    "<p>Live <strong>jazz</strong></p><p>All night</p>",
		Tickets: []models.TicketType{
			{Kind: models.TicketPaid, Price: 150000},
			{Kind: models.TicketStreaming, Price: 50000},
		},
	}

	document := NewDocument(&event, "Blue Note")
	if document.OrganizerName != "Blue Note" || document.CategorySlug != "music" || document.CategoryName != "Music" {
		t.Errorf("unexpected names: %+v", document)
	}
	if document.EndsAt != starts.Unix() {
		t.Errorf("EndsAt = %d, want the start %d when there is no end", document.EndsAt, starts.Unix())
	}
	if document.StartingPrice == nil || *document.StartingPrice != 50000 {
		t.Errorf("StartingPrice = %v, want 50000", document.StartingPrice)
	}
	if document.IsFree || !document.HasVenue || !document.HasStreaming {
		t.Errorf("flags = free %v venue %v streaming %v", document.IsFree, document.HasVenue, document.HasStreaming)
	}
	if document.Geo == nil || document.Geo.Lat != lat || document.Geo.Lng != lng {
		t.Errorf("Geo = %+v", document.Geo)
	}
	if document.Description != "Live jazz All night" {
		t.Errorf("Description = %q", document.Description)
	}
}

func TestNewDocumentFreeTicket(t *testing.T) {
	starts := time.Now().Add(time.Hour)
	event := models.Event{
		Id: "e2", Status: models.EventStatusPublished, StartsAt: &starts,
		Tickets: []models.TicketType{{Kind: models.TicketPaid, Price: 10000}, {Kind: models.TicketFree, Price: 999}},
	}
	document := NewDocument(&event, "")
	if !document.IsFree || document.StartingPrice == nil || *document.StartingPrice != 0 {
		t.Errorf("free ticket: IsFree %v StartingPrice %v", document.IsFree, document.StartingPrice)
	}
	if document.Geo != nil {
		t.Errorf("Geo should be omitted without a pin")
	}
}

func TestPlainTextLimit(t *testing.T) {
	if got := PlainText("<p>héllo &amp; world</p>", 7); got != "héllo &" {
		t.Errorf("PlainText = %q", got)
	}
}

func TestBackoff(t *testing.T) {
	if backoff(1) != 2*time.Second || backoff(3) != 8*time.Second || backoff(20) != maxBackoff {
		t.Errorf("backoff = %v %v %v", backoff(1), backoff(3), backoff(20))
	}
}
