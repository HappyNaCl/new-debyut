package seeders

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	eventmodels "github.com/HappyNaCl/new-debyut/internal/event/models"
	"gorm.io/gorm"
)

type sampleVenue struct {
	Name, Address, City string
	Latitude, Longitude float64
}

var sampleVenues = []sampleVenue{
	{"Istora Senayan", "Jl. Pintu Satu Senayan, Gelora", "Jakarta", -6.2185, 106.8023},
	{"JIExpo Kemayoran", "Jl. Benyamin Sueb, Pademangan Timur", "Jakarta", -6.1467, 106.8456},
	{"Ciputra Artpreneur", "Jl. Prof. Dr. Satrio Kav. 3-5, Kuningan", "Jakarta", -6.2246, 106.8287},
	{"Sabuga ITB", "Jl. Tamansari No. 73, Lb. Siliwangi", "Bandung", -6.8869, 107.6105},
	{"Braga Citywalk", "Jl. Braga No. 99-101, Braga", "Bandung", -6.9175, 107.6091},
	{"Grand City Convex", "Jl. Walikota Mustajab No. 1, Ketabang", "Surabaya", -7.2629, 112.7496},
	{"Jogja Expo Center", "Jl. Raya Janti, Banguntapan", "Yogyakarta", -7.7985, 110.4049},
	{"Taman Budaya Yogyakarta", "Jl. Sriwedani No. 1, Ngupasan", "Yogyakarta", -7.8003, 110.3676},
	{"Bali Nusa Dua Convention Center", "Kawasan Pariwisata Nusa Dua Lot NW/1", "Badung", -8.8011, 115.2283},
	{"Taman Werdhi Budaya", "Jl. Nusa Indah No. 1, Panjer", "Denpasar", -8.6573, 115.2337},
	{"Santika Premiere Dyandra", "Jl. Kapten Maulana Lubis No. 7", "Medan", 3.5903, 98.6741},
	{"Trans Studio Mall Makassar", "Jl. Metro Tanjung Bunga No. 1", "Makassar", -5.1554, 119.3949},
	{"Gedung Kesenian Gajayana", "Jl. Nusakambangan No. 19, Kasin", "Malang", -7.9861, 112.6283},
	{"Marina Convention Center", "Jl. Yos Sudarso, Tawangsari", "Semarang", -6.9524, 110.3936},
}

type sampleEvent struct {
	Category    string // slug
	Name        string
	Description string
	DurationHrs int
	// Price of the cheapest paid ticket in rupiah; 0 makes the event free.
	Price int64
	// Online events are sold only as STREAMING tickets and have no venue.
	Online bool
	// Hybrid events sell venue tickets and a streaming ticket.
	Hybrid bool
}

var sampleEvents = []sampleEvent{
	{"music", "Jazz Under the Stars", "An evening of smooth jazz from Indonesia's finest quartets, under an open sky.", 4, 250000, false, true},
	{"music", "Indie Rock Night Vol. 7", "Six local indie bands, one loud night. Standing room only.", 5, 150000, false, false},
	{"music", "Gamelan Orchestra Live", "A contemporary gamelan orchestra performs new works alongside classic Javanese pieces.", 2, 100000, false, false},
	{"music", "Acoustic Sunday Session", "Laid-back acoustic sets from singer-songwriters. Bring a friend.", 3, 0, false, false},
	{"music", "Electronic Music Festival", "Two stages, twelve DJs and a light show you won't forget.", 10, 450000, false, true},
	{"comedy-performance", "Stand-Up Comedy Special", "A full hour of brand-new material from headliners of the Jakarta comedy scene.", 2, 120000, false, true},
	{"comedy-performance", "Improv Theatre Night", "Every scene is made up on the spot from audience suggestions.", 2, 75000, false, false},
	{"comedy-performance", "Open Mic Comedy", "New comics try out their sets. Free entry, cheap laughs.", 3, 0, false, false},
	{"science-technology", "DevFest Indonesia", "Talks and codelabs on web, mobile, cloud and AI from community speakers.", 8, 200000, false, true},
	{"science-technology", "AI for Everyone Webinar", "A beginner-friendly introduction to machine learning, no maths required.", 2, 50000, true, false},
	{"science-technology", "48-Hour Hackathon", "Build something useful in two days. Mentors, meals and prizes included.", 48, 0, false, false},
	{"science-technology", "Cloud Native Meetup", "Kubernetes, observability and platform engineering war stories.", 3, 0, false, false},
	{"education", "Public Speaking Masterclass", "Practical techniques to speak with confidence, with live feedback.", 4, 350000, false, false},
	{"education", "Learn to Code Bootcamp Info Session", "Find out what a coding bootcamp is like and whether it's for you.", 1, 0, true, false},
	{"art-culture", "Contemporary Art Exhibition", "Paintings and installations from twenty emerging Indonesian artists.", 8, 50000, false, false},
	{"art-culture", "Batik Workshop", "Learn traditional hand-drawn batik and take your piece home.", 3, 175000, false, false},
	{"art-culture", "Wayang Kulit All Night", "A traditional shadow puppet performance, from dusk until dawn.", 8, 0, false, false},
	{"food-drink", "Street Food Festival", "Over 80 vendors serving the best street food from across the archipelago.", 9, 25000, false, false},
	{"food-drink", "Coffee Cupping Class", "Taste and compare single-origin Indonesian coffees with a Q grader.", 2, 150000, false, false},
	{"sports-fitness", "City Night Run 10K", "A 10K run through the city centre after dark. Medal for every finisher.", 4, 275000, false, false},
	{"sports-fitness", "Sunrise Yoga in the Park", "A gentle outdoor yoga flow for all levels. Mats provided.", 2, 0, false, false},
	{"sports-fitness", "Futsal Tournament", "Sixteen teams, one trophy. Register your squad or come to watch.", 10, 40000, false, false},
	{"business-finance", "Startup Pitch Night", "Ten early-stage startups pitch to investors. Networking after.", 3, 100000, false, true},
	{"business-finance", "Personal Finance 101", "Budgeting, saving and investing basics for young professionals.", 2, 75000, true, false},
	{"health-wellness", "Mindfulness Retreat Day", "A day of guided meditation, journaling and healthy food.", 7, 500000, false, false},
	{"family-kids", "Kids Science Fair", "Hands-on experiments and shows for children aged 5 to 12.", 6, 35000, false, false},
	{"fashion-beauty", "Modest Fashion Week", "Runway shows from local designers and a pop-up market.", 8, 125000, false, true},
	{"travel-outdoor", "Mount Bromo Sunrise Trip", "An overnight trip to watch the sunrise over Bromo. Transport included.", 20, 650000, false, false},
	{"media-entertainment", "Indie Film Screening and Q&A", "Award-winning Indonesian short films, followed by a talk with the directors.", 3, 60000, false, true},
	{"hobbies-lifestyle", "Board Game Marathon", "Hundreds of games to borrow and play. Snacks included.", 8, 50000, false, false},
}

// SeedEvents gives each organization that has no events yet perOrganization
// events: mostly upcoming and published, plus a few drafts and past events,
// across categories, cities, prices and venue, hybrid and online formats.
// Published events are queued in the search outbox so the event service
// indexes them. Banners come from picsum.photos.
func SeedEvents(eventDb *gorm.DB, organizations []SeededOrganization, perOrganization int) (int, error) {
	var categories []eventmodels.Category
	if err := eventDb.Find(&categories).Error; err != nil {
		return 0, err
	}
	categoryIds := make(map[string]string, len(categories))
	for _, category := range categories {
		categoryIds[category.Slug] = category.Id
	}

	created := 0
	for orgIndex, organization := range organizations {
		var count int64
		if err := eventDb.Model(&eventmodels.Event{}).Where("organization_id = ?", organization.Id).Count(&count).Error; err != nil {
			return created, err
		}
		if count > 0 {
			continue
		}

		// Seeded per organization, so a reseed after a reset looks the same.
		random := rand.New(rand.NewPCG(uint64(orgIndex)+1, 42))
		for index := range perOrganization {
			template := sampleEvents[random.IntN(len(sampleEvents))]
			event := buildSampleEvent(random, organization, template, categoryIds[template.Category], index)
			if err := insertSampleEvent(eventDb, event); err != nil {
				return created, fmt.Errorf("seed event %q: %w", event.Name, err)
			}
			created++
		}
	}
	return created, nil
}

func buildSampleEvent(random *rand.Rand, organization SeededOrganization, template sampleEvent, categoryId string, index int) *eventmodels.Event {
	now := time.Now().UTC()
	id := snowflake.Generate()

	// Every tenth event is a draft and every seventh one is over, so the
	// organizer dashboard tabs have something in them.
	status := eventmodels.EventStatusPublished
	days := 1 + random.IntN(120)
	switch {
	case index%10 == 9:
		status = eventmodels.EventStatusDraft
	case index%7 == 6:
		days = -(7 + random.IntN(60))
	}
	// Between 09:00 and 19:00 Jakarta time (UTC+7).
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startsAt := today.AddDate(0, 0, days).Add(time.Duration(2+random.IntN(11)) * time.Hour)
	endsAt := startsAt.Add(time.Duration(template.DurationHrs) * time.Hour)

	name := template.Name
	venue := sampleVenues[random.IntN(len(sampleVenues))]
	if !template.Online {
		name = fmt.Sprintf("%s %s", template.Name, venue.City)
	}

	event := &eventmodels.Event{
		Id:               id,
		OrganizationId:   organization.Id,
		CreatedBy:        organization.OwnerId,
		Status:           status,
		Name:             name,
		BannerUrl:        fmt.Sprintf("https://picsum.photos/seed/%s/1200/400", id),
		CoverUrl:         fmt.Sprintf("https://picsum.photos/seed/%s-cover/600/600", id),
		GalleryUrls:      []string{},
		StartsAt:         &startsAt,
		EndsAt:           &endsAt,
		Description:      sampleDescription(template),
		Terms:            "<p>Tickets are non-refundable. Bring a valid ID that matches the name on your ticket.</p>",
		ContactName:      "Event Team",
		ContactEmail:     "events@debyut.id",
		ContactPhone:     "+6281200000000",
		ServiceFeeBearer: eventmodels.ServiceFeeOrganizer,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if categoryId != "" {
		event.CategoryId = &categoryId
	}
	if !template.Online {
		event.VenueName = venue.Name
		event.Address = venue.Address
		event.City = venue.City
		lat, lng := venue.Latitude, venue.Longitude
		event.Latitude, event.Longitude = &lat, &lng
	}
	if status == eventmodels.EventStatusPublished {
		publishedAt := now.Add(-time.Duration(random.IntN(30*24)) * time.Hour)
		event.PublishedAt = &publishedAt
	}

	salesStart := now.AddDate(0, 0, -30)
	salesEnd := startsAt
	ticket := func(kind eventmodels.TicketKind, name string, price int64, quantity int32) eventmodels.TicketType {
		return eventmodels.TicketType{
			Id: snowflake.Generate(), EventId: id, Kind: kind, Name: name, Price: price, Quantity: quantity,
			SalesStartAt: &salesStart, SalesEndAt: &salesEnd, CreatedAt: now, UpdatedAt: now,
		}
	}
	switch {
	case template.Online:
		event.Tickets = append(event.Tickets, ticket(eventmodels.TicketStreaming, "Online Access", max(template.Price, 25000), 5000))
	case template.Price == 0:
		event.Tickets = append(event.Tickets, ticket(eventmodels.TicketFree, "Free Entry", 0, int32(100+random.IntN(400))))
	default:
		event.Tickets = append(event.Tickets,
			ticket(eventmodels.TicketPaid, "Regular", template.Price, int32(200+random.IntN(800))),
			ticket(eventmodels.TicketPaid, "VIP", template.Price*3, int32(20+random.IntN(80))),
		)
	}
	if template.Hybrid {
		event.Tickets = append(event.Tickets, ticket(eventmodels.TicketStreaming, "Livestream", max(template.Price/2, 25000), 2000))
	}
	for position := range event.Tickets {
		event.Tickets[position].Position = position
	}
	return event
}

func sampleDescription(template sampleEvent) string {
	var builder strings.Builder
	builder.WriteString("<p>" + template.Description + "</p>")
	builder.WriteString("<h3>What to expect</h3><ul>")
	if template.Online {
		builder.WriteString("<li>Watch live from anywhere, with a replay for 48 hours</li>")
	} else {
		builder.WriteString("<li>Doors open one hour before the start</li>")
	}
	builder.WriteString("<li>Food and drinks available</li><li>Suitable for all ages unless stated otherwise</li></ul>")
	return builder.String()
}

func insertSampleEvent(eventDb *gorm.DB, event *eventmodels.Event) error {
	return eventDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Tickets", "Category").Create(event).Error; err != nil {
			return err
		}
		if err := tx.Create(&event.Tickets).Error; err != nil {
			return err
		}
		if event.Status != eventmodels.EventStatusPublished {
			return nil
		}
		return tx.Create(&eventmodels.SearchOutbox{EventId: event.Id, AvailableAt: event.CreatedAt, CreatedAt: event.CreatedAt}).Error
	})
}
