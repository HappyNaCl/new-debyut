package seeders

import (
	"math/rand/v2"
	"testing"
	"time"

	eventmodels "github.com/HappyNaCl/new-debyut/internal/event/models"
	"github.com/HappyNaCl/new-debyut/internal/event/search"
)

func TestBuildSampleEvents(t *testing.T) {
	random := rand.New(rand.NewPCG(1, 42))
	organization := SeededOrganization{Id: "org", OwnerId: "owner"}
	now := time.Now()
	counts := map[string]int{}

	for index := range 70 {
		template := sampleEvents[random.IntN(len(sampleEvents))]
		event := buildSampleEvent(random, organization, template, "category", index)

		if event.CreatedBy != "owner" || event.OrganizationId != "org" {
			t.Fatalf("%s: wrong owner", event.Name)
		}
		if !event.EndsAt.After(*event.StartsAt) {
			t.Errorf("%s: ends before it starts", event.Name)
		}
		hasVenueTicket := false
		for _, ticket := range event.Tickets {
			if ticket.Kind != eventmodels.TicketFree && ticket.Price < 1 {
				t.Errorf("%s: %s ticket without a price", event.Name, ticket.Kind)
			}
			if ticket.SalesEndAt.After(*event.EndsAt) {
				t.Errorf("%s: sales end after the event", event.Name)
			}
			hasVenueTicket = hasVenueTicket || ticket.Kind != eventmodels.TicketStreaming
		}
		if hasVenueTicket && (event.Latitude == nil || event.VenueName == "" || event.City == "") {
			t.Errorf("%s: venue tickets but no venue", event.Name)
		}
		if (event.Status == eventmodels.EventStatusPublished) != (event.PublishedAt != nil) {
			t.Errorf("%s: published_at doesn't match status", event.Name)
		}

		switch {
		case event.Status == eventmodels.EventStatusDraft:
			counts["draft"]++
		case search.Indexable(event, now):
			counts["upcoming"]++
		default:
			counts["past"]++
		}
	}
	if counts["draft"] == 0 || counts["past"] == 0 || counts["upcoming"] < 50 {
		t.Errorf("unexpected mix: %v", counts)
	}
	t.Logf("mix: %v", counts)
}
