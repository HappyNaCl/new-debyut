package dto

import (
	"testing"

	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
)

func summaryOf(tickets ...*event.TicketType) *EventSummaryResponse {
	return NewEventSummaryResponse(&event.Event{Details: &event.EventInput{Tickets: tickets}}, &users.OrganizationProfile{Name: "Org"})
}

func TestSummaryStartingPrice(t *testing.T) {
	if got := summaryOf(); got.StartingPrice != nil || got.StreamingOnly {
		t.Fatalf("no tickets: got price %v, streamingOnly %v", got.StartingPrice, got.StreamingOnly)
	}
	got := summaryOf(&event.TicketType{Kind: "PAID", Price: 150000}, &event.TicketType{Kind: "STREAMING", Price: 50000})
	if got.StartingPrice == nil || *got.StartingPrice != 50000 || got.StreamingOnly {
		t.Fatalf("paid+streaming: got %v, streamingOnly %v", got.StartingPrice, got.StreamingOnly)
	}
	got = summaryOf(&event.TicketType{Kind: "PAID", Price: 150000}, &event.TicketType{Kind: "FREE", Price: 999})
	if *got.StartingPrice != 0 {
		t.Fatalf("free ticket should make it start from 0, got %d", *got.StartingPrice)
	}
	if got := summaryOf(&event.TicketType{Kind: "STREAMING", Price: 1}); !got.StreamingOnly {
		t.Fatal("streaming-only event not flagged")
	}
	if got := NewEventSummaryResponse(&event.Event{Details: &event.EventInput{}}, nil); got.Organizer.Name != "" {
		t.Fatal("missing organizer should give an empty name, not panic")
	}
}
