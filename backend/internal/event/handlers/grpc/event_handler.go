package handlers

import (
	"context"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/event/models"
	"github.com/HappyNaCl/new-debyut/internal/event/services"
	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *EventGrpcHandler) UploadEventImage(ctx context.Context, req *event.UploadEventImageRequest) (*event.UploadEventImageResponse, error) {
	url, err := h.eventSvc.UploadImage(ctx, req.GetOrganizationId(), services.ImageKind(req.GetKind()), req.GetContent())
	if err != nil {
		return nil, err
	}
	return &event.UploadEventImageResponse{Url: url}, nil
}

func (h *EventGrpcHandler) CreateEvent(ctx context.Context, req *event.SaveEventRequest) (*event.Event, error) {
	created, err := h.eventSvc.Create(ctx, req.GetOrganizationId(), req.GetActorId(), fromEventInput(req.GetEvent()), req.GetPublish())
	if err != nil {
		return nil, err
	}
	return toEventResponse(created), nil
}

func (h *EventGrpcHandler) UpdateEvent(ctx context.Context, req *event.SaveEventRequest) (*event.Event, error) {
	updated, err := h.eventSvc.UpdateDraft(ctx, req.GetOrganizationId(), req.GetEventId(), fromEventInput(req.GetEvent()), req.GetPublish())
	if err != nil {
		return nil, err
	}
	return toEventResponse(updated), nil
}

func (h *EventGrpcHandler) GetEvent(ctx context.Context, req *event.GetEventRequest) (*event.Event, error) {
	found, err := h.eventSvc.Get(ctx, req.GetOrganizationId(), req.GetEventId())
	if err != nil {
		return nil, err
	}
	return toEventResponse(found), nil
}

func (h *EventGrpcHandler) GetPublishedEvent(ctx context.Context, req *event.GetPublishedEventRequest) (*event.Event, error) {
	found, err := h.eventSvc.GetPublished(ctx, req.GetEventId())
	if err != nil {
		return nil, err
	}
	return toEventResponse(found), nil
}

func (h *EventGrpcHandler) ListLatestEvents(ctx context.Context, req *event.ListLatestEventsRequest) (*event.ListEventsResponse, error) {
	events, err := h.eventSvc.ListLatest(ctx, int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	response := &event.ListEventsResponse{Events: make([]*event.Event, 0, len(events))}
	for index := range events {
		response.Events = append(response.Events, toEventResponse(&events[index]))
	}
	return response, nil
}

func (h *EventGrpcHandler) ListEvents(ctx context.Context, req *event.ListEventsRequest) (*event.ListEventsResponse, error) {
	events, err := h.eventSvc.List(ctx, req.GetOrganizationId(), req.GetFilter())
	if err != nil {
		return nil, err
	}
	response := &event.ListEventsResponse{Events: make([]*event.Event, 0, len(events))}
	for index := range events {
		response.Events = append(response.Events, toEventResponse(&events[index]))
	}
	return response, nil
}

func fromEventInput(input *event.EventInput) *models.Event {
	result := &models.Event{
		Name:             input.GetName(),
		BannerUrl:        input.GetBannerUrl(),
		CoverUrl:         input.GetCoverUrl(),
		GalleryUrls:      input.GetGalleryUrls(),
		StartsAt:         fromTimestamp(input.GetStartsAt()),
		EndsAt:           fromTimestamp(input.GetEndsAt()),
		VenueName:        input.GetVenueName(),
		Address:          input.GetAddress(),
		City:             input.GetCity(),
		Latitude:         input.Latitude,
		Longitude:        input.Longitude,
		Description:      input.GetDescription(),
		Terms:            input.GetTerms(),
		ContactName:      input.GetContactName(),
		ContactEmail:     input.GetContactEmail(),
		ContactPhone:     input.GetContactPhone(),
		ServiceFeeBearer: models.ServiceFeeBearer(input.GetServiceFeeBearer()),
		Tickets:          make([]models.TicketType, 0, len(input.GetTickets())),
	}
	if categoryId := input.GetCategoryId(); categoryId != "" {
		result.CategoryId = &categoryId
	}
	for _, ticket := range input.GetTickets() {
		result.Tickets = append(result.Tickets, models.TicketType{
			Kind:         models.TicketKind(ticket.GetKind()),
			Name:         ticket.GetName(),
			Description:  ticket.GetDescription(),
			Price:        ticket.GetPrice(),
			Quantity:     ticket.GetQuantity(),
			SalesStartAt: fromTimestamp(ticket.GetSalesStartAt()),
			SalesEndAt:   fromTimestamp(ticket.GetSalesEndAt()),
		})
	}
	return result
}

func toEventResponse(source *models.Event) *event.Event {
	details := &event.EventInput{
		Name:             source.Name,
		BannerUrl:        source.BannerUrl,
		CoverUrl:         source.CoverUrl,
		GalleryUrls:      source.GalleryUrls,
		StartsAt:         toTimestamp(source.StartsAt),
		EndsAt:           toTimestamp(source.EndsAt),
		VenueName:        source.VenueName,
		Address:          source.Address,
		City:             source.City,
		Latitude:         source.Latitude,
		Longitude:        source.Longitude,
		Description:      source.Description,
		Terms:            source.Terms,
		ContactName:      source.ContactName,
		ContactEmail:     source.ContactEmail,
		ContactPhone:     source.ContactPhone,
		ServiceFeeBearer: string(source.ServiceFeeBearer),
		Tickets:          make([]*event.TicketType, 0, len(source.Tickets)),
	}
	if source.CategoryId != nil {
		details.CategoryId = *source.CategoryId
	}
	for index := range source.Tickets {
		ticket := &source.Tickets[index]
		details.Tickets = append(details.Tickets, &event.TicketType{
			Id:           ticket.Id,
			Kind:         string(ticket.Kind),
			Name:         ticket.Name,
			Description:  ticket.Description,
			Price:        ticket.Price,
			Quantity:     ticket.Quantity,
			SalesStartAt: toTimestamp(ticket.SalesStartAt),
			SalesEndAt:   toTimestamp(ticket.SalesEndAt),
		})
	}
	return &event.Event{
		Id:             source.Id,
		OrganizationId: source.OrganizationId,
		Status:         string(source.Status),
		Details:        details,
		CreatedBy:      source.CreatedBy,
		PublishedAt:    toTimestamp(source.PublishedAt),
		CreatedAt:      timestamppb.New(source.CreatedAt),
		UpdatedAt:      timestamppb.New(source.UpdatedAt),
	}
}

func fromTimestamp(value *timestamppb.Timestamp) *time.Time {
	if value == nil {
		return nil
	}
	converted := value.AsTime().UTC()
	return &converted
}

func toTimestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}
