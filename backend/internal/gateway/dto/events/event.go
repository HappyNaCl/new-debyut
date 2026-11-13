package dto

import (
	"time"

	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TicketInput is one ticket type. Price is in whole rupiah, ignored for FREE.
type TicketInput struct {
	Kind         string     `json:"kind"         binding:"required,oneof=PAID FREE STREAMING"`
	Name         string     `json:"name"         binding:"max=100"`
	Description  string     `json:"description"  binding:"max=500"`
	Price        int64      `json:"price"        binding:"min=0"`
	Quantity     int32      `json:"quantity"     binding:"min=0"`
	SalesStartAt *time.Time `json:"salesStartAt"`
	SalesEndAt   *time.Time `json:"salesEndAt"`
}

// SaveEventRequest creates or updates an event. With publish false it is
// saved as a draft (only the name is required); with publish true every
// field needed to sell tickets must be present.
type SaveEventRequest struct {
	Publish          bool          `json:"publish"`
	Name             string        `json:"name"             binding:"max=150"`
	CategoryId       string        `json:"categoryId"       binding:"max=20"`
	BannerUrl        string        `json:"bannerUrl"        binding:"max=512"`
	CoverUrl         string        `json:"coverUrl"         binding:"max=512"`
	GalleryUrls      []string      `json:"galleryUrls"      binding:"max=6,dive,max=512"`
	StartsAt         *time.Time    `json:"startsAt"`
	EndsAt           *time.Time    `json:"endsAt"`
	VenueName        string        `json:"venueName"        binding:"max=150"`
	Address          string        `json:"address"          binding:"max=500"`
	City             string        `json:"city"             binding:"max=100"`
	Latitude         *float64      `json:"latitude"         binding:"omitempty,min=-90,max=90"`
	Longitude        *float64      `json:"longitude"        binding:"omitempty,min=-180,max=180"`
	Description      string        `json:"description"      binding:"max=200000"` // HTML, sanitized server-side
	Terms            string        `json:"terms"            binding:"max=200000"` // HTML, sanitized server-side
	ContactName      string        `json:"contactName"      binding:"max=100"`
	ContactEmail     string        `json:"contactEmail"     binding:"max=255"`
	ContactPhone     string        `json:"contactPhone"     binding:"max=20"`
	ServiceFeeBearer string        `json:"serviceFeeBearer" binding:"omitempty,oneof=ORGANIZER BUYER"`
	Tickets          []TicketInput `json:"tickets"          binding:"max=20,dive"`
}

type ListEventsQuery struct {
	Filter string `form:"filter" binding:"required,oneof=DRAFT ACTIVE PAST"`
}

type ImageResponse struct {
	Url string `json:"url"`
}

type TicketResponse struct {
	Id           string     `json:"id"`
	Kind         string     `json:"kind"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Price        int64      `json:"price"`
	Quantity     int32      `json:"quantity"`
	SalesStartAt *time.Time `json:"salesStartAt"`
	SalesEndAt   *time.Time `json:"salesEndAt"`
}

type EventResponse struct {
	Id               string           `json:"id"`
	OrganizationId   string           `json:"organizationId"`
	Status           string           `json:"status"`
	Name             string           `json:"name"`
	CategoryId       string           `json:"categoryId"`
	BannerUrl        string           `json:"bannerUrl"`
	CoverUrl         string           `json:"coverUrl"`
	GalleryUrls      []string         `json:"galleryUrls"`
	StartsAt         *time.Time       `json:"startsAt"`
	EndsAt           *time.Time       `json:"endsAt"`
	VenueName        string           `json:"venueName"`
	Address          string           `json:"address"`
	City             string           `json:"city"`
	Latitude         *float64         `json:"latitude"`
	Longitude        *float64         `json:"longitude"`
	Description      string           `json:"description"`
	Terms            string           `json:"terms"`
	ContactName      string           `json:"contactName"`
	ContactEmail     string           `json:"contactEmail"`
	ContactPhone     string           `json:"contactPhone"`
	ServiceFeeBearer string           `json:"serviceFeeBearer"`
	Tickets          []TicketResponse `json:"tickets"`
	PublishedAt      *time.Time       `json:"publishedAt"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

func (r *SaveEventRequest) ToInput() *event.EventInput {
	input := &event.EventInput{
		Name:             r.Name,
		CategoryId:       r.CategoryId,
		BannerUrl:        r.BannerUrl,
		CoverUrl:         r.CoverUrl,
		GalleryUrls:      r.GalleryUrls,
		StartsAt:         timestamp(r.StartsAt),
		EndsAt:           timestamp(r.EndsAt),
		VenueName:        r.VenueName,
		Address:          r.Address,
		City:             r.City,
		Latitude:         r.Latitude,
		Longitude:        r.Longitude,
		Description:      r.Description,
		Terms:            r.Terms,
		ContactName:      r.ContactName,
		ContactEmail:     r.ContactEmail,
		ContactPhone:     r.ContactPhone,
		ServiceFeeBearer: r.ServiceFeeBearer,
		Tickets:          make([]*event.TicketType, 0, len(r.Tickets)),
	}
	for _, ticket := range r.Tickets {
		input.Tickets = append(input.Tickets, &event.TicketType{
			Kind:         ticket.Kind,
			Name:         ticket.Name,
			Description:  ticket.Description,
			Price:        ticket.Price,
			Quantity:     ticket.Quantity,
			SalesStartAt: timestamp(ticket.SalesStartAt),
			SalesEndAt:   timestamp(ticket.SalesEndAt),
		})
	}
	return input
}

func NewEventResponse(res *event.Event) *EventResponse {
	details := res.GetDetails()
	response := &EventResponse{
		Id:               res.GetId(),
		OrganizationId:   res.GetOrganizationId(),
		Status:           res.GetStatus(),
		Name:             details.GetName(),
		CategoryId:       details.GetCategoryId(),
		BannerUrl:        details.GetBannerUrl(),
		CoverUrl:         details.GetCoverUrl(),
		GalleryUrls:      details.GetGalleryUrls(),
		StartsAt:         timeOf(details.GetStartsAt()),
		EndsAt:           timeOf(details.GetEndsAt()),
		VenueName:        details.GetVenueName(),
		Address:          details.GetAddress(),
		City:             details.GetCity(),
		Latitude:         details.Latitude,
		Longitude:        details.Longitude,
		Description:      details.GetDescription(),
		Terms:            details.GetTerms(),
		ContactName:      details.GetContactName(),
		ContactEmail:     details.GetContactEmail(),
		ContactPhone:     details.GetContactPhone(),
		ServiceFeeBearer: details.GetServiceFeeBearer(),
		Tickets:          make([]TicketResponse, 0, len(details.GetTickets())),
		PublishedAt:      timeOf(res.GetPublishedAt()),
		CreatedAt:        res.GetCreatedAt().AsTime(),
		UpdatedAt:        res.GetUpdatedAt().AsTime(),
	}
	if response.GalleryUrls == nil {
		response.GalleryUrls = []string{}
	}
	for _, ticket := range details.GetTickets() {
		response.Tickets = append(response.Tickets, TicketResponse{
			Id:           ticket.GetId(),
			Kind:         ticket.GetKind(),
			Name:         ticket.GetName(),
			Description:  ticket.GetDescription(),
			Price:        ticket.GetPrice(),
			Quantity:     ticket.GetQuantity(),
			SalesStartAt: timeOf(ticket.GetSalesStartAt()),
			SalesEndAt:   timeOf(ticket.GetSalesEndAt()),
		})
	}
	return response
}

func NewEventListResponse(res *event.ListEventsResponse) []*EventResponse {
	response := make([]*EventResponse, 0, len(res.GetEvents()))
	for _, item := range res.GetEvents() {
		response = append(response, NewEventResponse(item))
	}
	return response
}

func timestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}

func timeOf(value *timestamppb.Timestamp) *time.Time {
	if value == nil {
		return nil
	}
	converted := value.AsTime()
	return &converted
}

type OrganizerResponse struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatarUrl"`
}

// PublicEventResponse is what anyone can see on an event page. It leaves out
// organizer-only details (contact info, who created it, draft timestamps).
type PublicEventResponse struct {
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	CategoryId       string            `json:"categoryId"`
	BannerUrl        string            `json:"bannerUrl"`
	CoverUrl         string            `json:"coverUrl"`
	GalleryUrls      []string          `json:"galleryUrls"`
	StartsAt         *time.Time        `json:"startsAt"`
	EndsAt           *time.Time        `json:"endsAt"`
	VenueName        string            `json:"venueName"`
	Address          string            `json:"address"`
	City             string            `json:"city"`
	Latitude         *float64          `json:"latitude"`
	Longitude        *float64          `json:"longitude"`
	Description      string            `json:"description"`
	Terms            string            `json:"terms"`
	ServiceFeeBearer string            `json:"serviceFeeBearer"`
	Tickets          []TicketResponse  `json:"tickets"`
	Organizer        OrganizerResponse `json:"organizer"`
	PublishedAt      *time.Time        `json:"publishedAt"`
}

func NewPublicEventResponse(res *event.Event, organizer *users.OrganizationProfile) *PublicEventResponse {
	full := NewEventResponse(res)
	return &PublicEventResponse{
		Id:               full.Id,
		Name:             full.Name,
		CategoryId:       full.CategoryId,
		BannerUrl:        full.BannerUrl,
		CoverUrl:         full.CoverUrl,
		GalleryUrls:      full.GalleryUrls,
		StartsAt:         full.StartsAt,
		EndsAt:           full.EndsAt,
		VenueName:        full.VenueName,
		Address:          full.Address,
		City:             full.City,
		Latitude:         full.Latitude,
		Longitude:        full.Longitude,
		Description:      full.Description,
		Terms:            full.Terms,
		ServiceFeeBearer: full.ServiceFeeBearer,
		Tickets:          full.Tickets,
		Organizer: OrganizerResponse{
			Id:        organizer.GetId(),
			Name:      organizer.GetName(),
			AvatarUrl: organizer.GetAvatarUrl(),
		},
		PublishedAt: full.PublishedAt,
	}
}

type ListLatestEventsQuery struct {
	Limit int32 `form:"limit" binding:"omitempty,min=1,max=50"`
}

// EventSummaryResponse is an event card: enough to list it, without the
// description, terms or full ticket details.
type EventSummaryResponse struct {
	Id            string     `json:"id"`
	Name          string     `json:"name"`
	CategoryId    string     `json:"categoryId"`
	BannerUrl     string     `json:"bannerUrl"`
	StartsAt      *time.Time `json:"startsAt"`
	EndsAt        *time.Time `json:"endsAt"`
	VenueName     string     `json:"venueName"`
	City          string     `json:"city"`
	StreamingOnly bool       `json:"streamingOnly"`
	// Lowest ticket price in rupiah (0 when a free ticket exists); null when there are no tickets.
	StartingPrice *int64            `json:"startingPrice"`
	Organizer     OrganizerResponse `json:"organizer"`
	PublishedAt   *time.Time        `json:"publishedAt"`
}

func NewEventSummaryResponse(res *event.Event, organizer *users.OrganizationProfile) *EventSummaryResponse {
	details := res.GetDetails()
	summary := &EventSummaryResponse{
		Id:          res.GetId(),
		Name:        details.GetName(),
		CategoryId:  details.GetCategoryId(),
		BannerUrl:   details.GetBannerUrl(),
		StartsAt:    timeOf(details.GetStartsAt()),
		EndsAt:      timeOf(details.GetEndsAt()),
		VenueName:   details.GetVenueName(),
		City:        details.GetCity(),
		PublishedAt: timeOf(res.GetPublishedAt()),
		Organizer: OrganizerResponse{
			Id:        organizer.GetId(),
			Name:      organizer.GetName(),
			AvatarUrl: organizer.GetAvatarUrl(),
		},
	}
	summary.StreamingOnly = len(details.GetTickets()) > 0
	for _, ticket := range details.GetTickets() {
		price := ticket.GetPrice()
		if ticket.GetKind() == "FREE" {
			price = 0
		}
		if summary.StartingPrice == nil || price < *summary.StartingPrice {
			summary.StartingPrice = &price
		}
		if ticket.GetKind() != "STREAMING" {
			summary.StreamingOnly = false
		}
	}
	return summary
}
