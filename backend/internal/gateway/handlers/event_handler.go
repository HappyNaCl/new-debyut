package handlers

import (
	"net/http"

	"github.com/HappyNaCl/new-debyut/internal/common/validate"
	dto "github.com/HappyNaCl/new-debyut/internal/gateway/dto/events"
	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/gin-gonic/gin"
)

const maxEventImageBytes = 2 << 20

type EventHttpHandler struct {
	eventClient        event.EventServiceClient
	organizationClient users.OrganizationServiceClient
}

func NewEventHttpHandler(eventClient event.EventServiceClient, organizationClient users.OrganizationServiceClient) *EventHttpHandler {
	return &EventHttpHandler{eventClient: eventClient, organizationClient: organizationClient}
}

// requireMember lets the request through only for active members of the
// organization in the path; the users service answers 404 for anyone else.
func (h EventHttpHandler) requireMember(c *gin.Context) bool {
	if _, err := h.organizationClient.GetOrganization(c, &users.OrganizationRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
	}); err != nil {
		respondError(c, err)
		return false
	}
	return true
}

// @Summary      Upload an event image
// @Description  Uploads a banner (up to 2 MB), 1:1 sales-page cover (up to 2 MB) or gallery photo (up to 1 MB) and returns its URL, to be referenced when saving the event. JPG, PNG or WebP.
// @Tags         Events
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      string  true  "Organization ID"
// @Param        kind   formData  string  true  "Image kind"  Enums(BANNER, COVER, GALLERY)
// @Param        image  formData  file    true  "Image file"
// @Success      201  {object}  object{data=dto.ImageResponse}
// @Failure      400  {object}  object{error=object}  "Missing, too large, or unsupported image"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      404  {object}  object{error=string}  "Organization not found"
// @Router       /organizations/{id}/events/images [post]
func (h EventHttpHandler) UploadImage(c *gin.Context) {
	if !h.requireMember(c) {
		return
	}
	content, ok := readImage(c, "image", maxEventImageBytes)
	if !ok {
		return
	}

	res, err := h.eventClient.UploadEventImage(c, &event.UploadEventImageRequest{
		OrganizationId: c.Param("id"),
		Kind:           c.PostForm("kind"),
		Content:        content,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dto.ImageResponse{Url: res.GetUrl()}})
}

// @Summary      List an organization's events
// @Description  DRAFT: unpublished drafts. ACTIVE: published and not yet ended. PAST: published and ended.
// @Tags         Events
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "Organization ID"
// @Param        filter  query     string  true  "Which events"  Enums(DRAFT, ACTIVE, PAST)
// @Success      200  {object}  object{data=[]dto.EventResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      404  {object}  object{error=string}  "Organization not found"
// @Router       /organizations/{id}/events [get]
func (h EventHttpHandler) List(c *gin.Context) {
	var query dto.ListEventsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}
	if !h.requireMember(c) {
		return
	}

	res, err := h.eventClient.ListEvents(c, &event.ListEventsRequest{OrganizationId: c.Param("id"), Filter: query.Filter})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewEventListResponse(res)})
}

// @Summary      Create an event
// @Description  Saves a draft (publish=false; only the name is required) or publishes it (publish=true; every field needed to sell tickets is required). Problems come back as field errors, e.g. {"error": {"name": "...", "tickets.0.price": "..."}}.
// @Tags         Events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                true  "Organization ID"
// @Param        request  body      dto.SaveEventRequest  true  "Event"
// @Success      201  {object}  object{data=dto.EventResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      404  {object}  object{error=string}  "Organization not found"
// @Router       /organizations/{id}/events [post]
func (h EventHttpHandler) Create(c *gin.Context) {
	var req dto.SaveEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}
	if !h.requireMember(c) {
		return
	}

	res, err := h.eventClient.CreateEvent(c, &event.SaveEventRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
		Event:          req.ToInput(),
		Publish:        req.Publish,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dto.NewEventResponse(res)})
}

// @Summary      Get an event
// @Tags         Events
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true  "Organization ID"
// @Param        eventId  path      string  true  "Event ID"
// @Success      200  {object}  object{data=dto.EventResponse}
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      404  {object}  object{error=string}  "Organization or event not found"
// @Router       /organizations/{id}/events/{eventId} [get]
func (h EventHttpHandler) Get(c *gin.Context) {
	if !h.requireMember(c) {
		return
	}

	res, err := h.eventClient.GetEvent(c, &event.GetEventRequest{OrganizationId: c.Param("id"), EventId: c.Param("eventId")})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewEventResponse(res)})
}

// @Summary      Update a draft event
// @Description  Replaces the draft's details and tickets. Send publish=true to publish it. Published events can't be edited yet.
// @Tags         Events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                true  "Organization ID"
// @Param        eventId  path      string                true  "Event ID"
// @Param        request  body      dto.SaveEventRequest  true  "Event"
// @Success      200  {object}  object{data=dto.EventResponse}
// @Failure      400  {object}  object{error=object}  "Validation error, or the event is already published"
// @Failure      401  {object}  object{error=string}  "Missing or invalid access token"
// @Failure      404  {object}  object{error=string}  "Organization or event not found"
// @Router       /organizations/{id}/events/{eventId} [put]
func (h EventHttpHandler) Update(c *gin.Context) {
	var req dto.SaveEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}
	if !h.requireMember(c) {
		return
	}

	res, err := h.eventClient.UpdateEvent(c, &event.SaveEventRequest{
		OrganizationId: c.Param("id"),
		ActorId:        c.GetString("userId"),
		EventId:        c.Param("eventId"),
		Event:          req.ToInput(),
		Publish:        req.Publish,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewEventResponse(res)})
}

// @Summary      Get a published event (public)
// @Description  Anyone can view a published event; drafts are not found. Description and terms are sanitized HTML. Organizer contact details are not included.
// @Tags         Events
// @Produce      json
// @Param        id   path      string  true  "Event ID"
// @Success      200  {object}  object{data=dto.PublicEventResponse}
// @Failure      404  {object}  object{error=string}  "Event not found"
// @Router       /events/{id} [get]
func (h EventHttpHandler) GetPublic(c *gin.Context) {
	res, err := h.eventClient.GetPublishedEvent(c, &event.GetPublishedEventRequest{EventId: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}

	organizer, err := h.organizationClient.GetOrganizationProfile(c, &users.OrganizationProfileRequest{
		OrganizationId: res.GetOrganizationId(),
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.NewPublicEventResponse(res, organizer)})
}

// @Summary      List the latest events (public)
// @Description  The most recently published events that haven't ended, newest first.
// @Tags         Events
// @Produce      json
// @Param        limit  query     int  false  "How many (1-50, default 8)"
// @Success      200  {object}  object{data=[]dto.EventSummaryResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Router       /events [get]
func (h EventHttpHandler) ListLatest(c *gin.Context) {
	var query dto.ListLatestEventsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}

	res, err := h.eventClient.ListLatestEvents(c, &event.ListLatestEventsRequest{Limit: query.Limit})
	if err != nil {
		respondError(c, err)
		return
	}

	summaries, err := h.summaries(c, res.GetEvents())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summaries})
}

// @Summary      Search events (public)
// @Description  Full-text search over published events that haven't ended, with filters. Without q, results are sorted soonest first. cities and categories count matches per city and category slug across all pages.
// @Tags         Events
// @Produce      json
// @Param        q         query     string  false  "Search text"
// @Param        category  query     string  false  "Category slug"
// @Param        city      query     string  false  "City (exact)"
// @Param        from      query     string  false  "RFC 3339: still running at or after"
// @Param        to        query     string  false  "RFC 3339: starting at or before"
// @Param        format    query     string  false  "Venue or online"  Enums(VENUE, ONLINE)
// @Param        free      query     bool    false  "Only events with a free ticket"
// @Param        lat       query     number  false  "Latitude (with lng and radiusKm)"
// @Param        lng       query     number  false  "Longitude"
// @Param        radiusKm  query     number  false  "Radius in km (up to 500)"
// @Param        sort      query     string  false  "Order"  Enums(RELEVANCE, SOONEST, NEWEST, PRICE_LOW)
// @Param        page      query     int     false  "Page (default 1)"
// @Param        pageSize  query     int     false  "Page size (1-48, default 24)"
// @Success      200  {object}  object{data=dto.SearchEventsResponse}
// @Failure      400  {object}  object{error=object}  "Validation error"
// @Failure      503  {object}  object{error=string}  "Search is unavailable"
// @Router       /events/search [get]
func (h EventHttpHandler) Search(c *gin.Context) {
	var query dto.SearchEventsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validate.ParseErrors(err)})
		return
	}
	near := []bool{query.Latitude != nil, query.Longitude != nil, query.RadiusKm != nil}
	if (near[0] || near[1] || near[2]) && !(near[0] && near[1] && near[2]) {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"radiusKm": "lat, lng and radiusKm must be sent together."}})
		return
	}

	res, err := h.eventClient.SearchEvents(c, query.ToRequest())
	if err != nil {
		respondError(c, err)
		return
	}

	summaries, err := h.summaries(c, res.GetEvents())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dto.SearchEventsResponse{
		Events:     summaries,
		Total:      res.GetTotal(),
		Page:       res.GetPage(),
		PageSize:   res.GetPageSize(),
		Cities:     res.GetCities(),
		Categories: res.GetCategories(),
	}})
}

// summaries turns events into cards, looking their organizers up in one call.
func (h EventHttpHandler) summaries(c *gin.Context, events []*event.Event) ([]*dto.EventSummaryResponse, error) {
	summaries := make([]*dto.EventSummaryResponse, 0, len(events))
	if len(events) == 0 {
		return summaries, nil
	}

	organizationIds := make([]string, 0, len(events))
	seen := map[string]bool{}
	for _, item := range events {
		if id := item.GetOrganizationId(); !seen[id] {
			seen[id] = true
			organizationIds = append(organizationIds, id)
		}
	}
	profiles, err := h.organizationClient.ListOrganizationProfiles(c, &users.ListOrganizationProfilesRequest{OrganizationIds: organizationIds})
	if err != nil {
		return nil, err
	}
	byId := make(map[string]*users.OrganizationProfile, len(profiles.GetOrganizations()))
	for _, profile := range profiles.GetOrganizations() {
		byId[profile.GetId()] = profile
	}

	for _, item := range events {
		summaries = append(summaries, dto.NewEventSummaryResponse(item, byId[item.GetOrganizationId()]))
	}
	return summaries, nil
}
