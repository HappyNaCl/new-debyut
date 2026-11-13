package services

import (
	"testing"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/event/models"
)

func completeEvent() *models.Event {
	categoryId := "1"
	start := time.Now().Add(48 * time.Hour)
	end := start.Add(3 * time.Hour)
	salesStart := time.Now()
	salesEnd := start
	lat, lng := -6.2, 106.8
	return &models.Event{
		Name:         "Jazz night",
		CategoryId:   &categoryId,
		BannerUrl:    "https://cdn.example/events/1/banner/a.jpg",
		StartsAt:     &start,
		EndsAt:       &end,
		Latitude:     &lat,
		Longitude:    &lng,
		VenueName:    "Hall",
		Address:      "Jl. Example 1",
		City:         "Jakarta",
		Description:  "An evening of jazz.",
		ContactName:  "Ana",
		ContactEmail: "ana@example.com",
		ContactPhone: "+628123456789",
		Tickets: []models.TicketType{{
			Kind: models.TicketPaid, Name: "Regular", Price: 50000, Quantity: 100,
			SalesStartAt: &salesStart, SalesEndAt: &salesEnd,
		}},
	}
}

func TestValidateCompleteAcceptsCompleteEvent(t *testing.T) {
	fields := map[string]string{}
	validateComplete(completeEvent(), fields)
	if len(fields) != 0 {
		t.Fatalf("expected no errors, got %v", fields)
	}
}

func TestValidateCompleteReportsMissingFields(t *testing.T) {
	event := completeEvent()
	event.CategoryId = nil
	event.BannerUrl = ""
	event.City = ""
	event.ContactPhone = ""
	event.Tickets[0].SalesEndAt = nil

	fields := map[string]string{}
	validateComplete(event, fields)
	for _, field := range []string{"categoryId", "bannerUrl", "city", "contactPhone", "tickets.0.salesEndAt"} {
		if _, ok := fields[field]; !ok {
			t.Errorf("expected an error for %s, got %v", field, fields)
		}
	}
}

func TestValidateCompleteRequiresTicketsAndFutureStart(t *testing.T) {
	event := completeEvent()
	past := time.Now().Add(-time.Hour)
	event.StartsAt = &past
	event.Tickets = nil

	fields := map[string]string{}
	validateComplete(event, fields)
	if fields["startsAt"] == "" || fields["tickets"] == "" {
		t.Fatalf("expected startsAt and tickets errors, got %v", fields)
	}
}

func TestValidateCompleteVenueRequiredOnlyForVenueTickets(t *testing.T) {
	event := completeEvent()
	event.VenueName, event.Address, event.City = "", "", ""
	event.Latitude, event.Longitude = nil, nil

	fields := map[string]string{}
	validateComplete(event, fields)
	for _, field := range []string{"location", "venueName", "address", "city"} {
		if fields[field] == "" {
			t.Errorf("expected %s error for a venue ticket, got %v", field, fields)
		}
	}

	event.Tickets[0].Kind = models.TicketStreaming
	fields = map[string]string{}
	validateComplete(event, fields)
	if len(fields) != 0 {
		t.Fatalf("streaming-only event should not need a venue, got %v", fields)
	}
}

func TestStreamingTicketsMustBePaid(t *testing.T) {
	event := completeEvent()
	event.Tickets[0].Kind = models.TicketStreaming
	event.Tickets[0].Price = 0

	fields := validateTickets(event)
	if fields["tickets.0.price"] == "" {
		t.Fatalf("expected a price error for a free streaming ticket, got %v", fields)
	}

	event.Tickets[0].Price = 25000
	if fields := validateTickets(event); len(fields) != 0 {
		t.Fatalf("expected a paid streaming ticket to be valid, got %v", fields)
	}
}
