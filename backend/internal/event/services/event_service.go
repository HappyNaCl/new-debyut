package services

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/common/emails"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	"github.com/HappyNaCl/new-debyut/internal/common/storage"
	"github.com/HappyNaCl/new-debyut/internal/event/models"
	"github.com/HappyNaCl/new-debyut/internal/event/repositories"
	"github.com/HappyNaCl/new-debyut/internal/event/search"
	"go.uber.org/zap"
)

const (
	maxGalleryImages  = 6
	maxTicketTypes    = 20
	maxTicketQuantity = 100000
	maxTicketPrice    = 1_000_000_000
	// Limits on rich text: visible characters, and the stored HTML size.
	maxRichTextLength = 20000
	maxRichTextBytes  = 100000
)

// ImageKind is what an uploaded event image is for; each kind has its own size
// limit, mirrored by the frontend.
type ImageKind string

const (
	ImageBanner  ImageKind = "BANNER"
	ImageCover   ImageKind = "COVER"
	ImageGallery ImageKind = "GALLERY"
)

var imageLimits = map[ImageKind]int{
	ImageBanner:  2 << 20,
	ImageCover:   2 << 20,
	ImageGallery: 1 << 20,
}

var imageExtensions = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

type EventService struct {
	eventRepository    *repositories.EventRepository
	categoryRepository *repositories.CategoryRepository
	storage            *storage.Client
	searcher           *search.Searcher
}

func NewEventService(
	eventRepository *repositories.EventRepository,
	categoryRepository *repositories.CategoryRepository,
	storage *storage.Client,
	searcher *search.Searcher,
) *EventService {
	return &EventService{
		eventRepository:    eventRepository,
		categoryRepository: categoryRepository,
		storage:            storage,
		searcher:           searcher,
	}
}

// UploadImage stores an event image and returns its public URL. Images are
// uploaded before the event is saved, so they are keyed by organization; the
// event later references them by URL.
func (s *EventService) UploadImage(ctx context.Context, organizationId string, kind ImageKind, content []byte) (string, error) {
	limit, ok := imageLimits[kind]
	if !ok {
		return "", apperrors.ErrInvalidImageKind
	}
	if len(content) == 0 || len(content) > limit {
		return "", apperrors.ErrInvalidEventImage
	}
	contentType := http.DetectContentType(content)
	extension, ok := imageExtensions[contentType]
	if !ok {
		return "", apperrors.ErrInvalidEventImage
	}

	key := fmt.Sprintf("%s%s/%s.%s", imagePrefix(organizationId), strings.ToLower(string(kind)), snowflake.Generate(), extension)
	url, err := s.storage.Upload(ctx, key, content, contentType)
	if err != nil {
		l.Logger.Error("failed to upload event image", zap.String("key", key), zap.Error(err))
		return "", apperrors.ErrEventImageUpload
	}
	return url, nil
}

func (s *EventService) Create(ctx context.Context, organizationId, actorId string, event *models.Event, publish bool) (*models.Event, error) {
	event.Id = snowflake.Generate()
	event.OrganizationId = organizationId
	event.CreatedBy = actorId
	if err := s.prepare(ctx, event, publish); err != nil {
		return nil, err
	}
	if err := s.eventRepository.Create(ctx, event); err != nil {
		return nil, err
	}
	return s.eventRepository.Get(ctx, organizationId, event.Id)
}

// UpdateDraft saves changes to a draft, optionally publishing it.
func (s *EventService) UpdateDraft(ctx context.Context, organizationId, eventId string, event *models.Event, publish bool) (*models.Event, error) {
	event.Id = eventId
	event.OrganizationId = organizationId
	if err := s.prepare(ctx, event, publish); err != nil {
		return nil, err
	}
	if err := s.eventRepository.UpdateDraft(ctx, event); err != nil {
		return nil, err
	}
	return s.eventRepository.Get(ctx, organizationId, eventId)
}

func (s *EventService) Get(ctx context.Context, organizationId, eventId string) (*models.Event, error) {
	return s.eventRepository.Get(ctx, organizationId, eventId)
}

func (s *EventService) GetPublished(ctx context.Context, eventId string) (*models.Event, error) {
	return s.eventRepository.GetPublished(ctx, eventId)
}

const (
	defaultLatestEvents = 8
	maxLatestEvents     = 50
)

func (s *EventService) ListLatest(ctx context.Context, limit int) ([]models.Event, error) {
	if limit < 1 {
		limit = defaultLatestEvents
	}
	return s.eventRepository.ListLatest(ctx, min(limit, maxLatestEvents))
}

const (
	defaultSearchPageSize = 24
	maxSearchPageSize     = 48
	maxSearchRadiusKm     = 500
	maxSearchQueryRunes   = 200
)

// SearchResult is one page of search results, ranked.
type SearchResult struct {
	Events     []models.Event
	Total      int64
	Page       int
	PageSize   int
	Cities     map[string]int
	Categories map[string]int
}

// Search finds published events that haven't ended. Meilisearch ranks the
// matches; the events themselves are read from the database, so a result
// whose index entry is briefly stale is dropped rather than shown wrong.
func (s *EventService) Search(ctx context.Context, query search.Query) (*SearchResult, error) {
	query.Text = strings.TrimSpace(query.Text)
	query.CategorySlug = strings.TrimSpace(query.CategorySlug)
	query.City = strings.TrimSpace(query.City)
	if utf8.RuneCountInString(query.Text) > maxSearchQueryRunes {
		return nil, apperrors.ErrInvalidEventSearch
	}
	switch query.Sort {
	case "":
		query.Sort = search.SortRelevance
	case search.SortRelevance, search.SortSoonest, search.SortNewest, search.SortPriceLow:
	default:
		return nil, apperrors.ErrInvalidEventSearch
	}
	switch query.Format {
	case search.FormatAny, search.FormatVenue, search.FormatOnline:
	default:
		return nil, apperrors.ErrInvalidEventSearch
	}
	if query.From != nil && query.To != nil && query.To.Before(*query.From) {
		return nil, apperrors.ErrInvalidEventSearch
	}
	if query.Near && (query.Latitude < -90 || query.Latitude > 90 || query.Longitude < -180 || query.Longitude > 180 ||
		query.RadiusKm <= 0 || query.RadiusKm > maxSearchRadiusKm) {
		return nil, apperrors.ErrInvalidEventSearch
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = defaultSearchPageSize
	}
	query.PageSize = min(query.PageSize, maxSearchPageSize)

	found, err := s.searcher.Search(ctx, query)
	if err != nil {
		l.Logger.Warn("event search failed", zap.Error(err))
		return nil, apperrors.ErrSearchUnavailable
	}
	events, err := s.eventRepository.ListUpcomingByIds(ctx, found.Ids)
	if err != nil {
		return nil, err
	}
	return &SearchResult{
		Events:     events,
		Total:      found.Total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		Cities:     found.Cities,
		Categories: found.Categories,
	}, nil
}

func (s *EventService) List(ctx context.Context, organizationId string, filter string) ([]models.Event, error) {
	return s.eventRepository.List(ctx, organizationId, repositories.EventFilter(filter))
}

// prepare normalizes the event, validates it (fully when publishing) and sets
// its status and ticket ids.
func (s *EventService) prepare(ctx context.Context, event *models.Event, publish bool) error {
	normalize(event)

	fields := s.validateShape(ctx, event)
	if publish {
		validateComplete(event, fields)
	}
	if len(fields) > 0 {
		return apperrors.EventValidationError(fields)
	}

	if event.ServiceFeeBearer == "" {
		event.ServiceFeeBearer = models.ServiceFeeOrganizer
	}
	event.Status = models.EventStatusDraft
	event.PublishedAt = nil
	if publish {
		now := time.Now().UTC()
		event.Status = models.EventStatusPublished
		event.PublishedAt = &now
	}
	for index := range event.Tickets {
		event.Tickets[index].Id = snowflake.Generate()
		if event.Tickets[index].Kind == models.TicketFree {
			event.Tickets[index].Price = 0
		}
	}
	return nil
}

func normalize(event *models.Event) {
	event.Name = strings.TrimSpace(event.Name)
	event.VenueName = strings.TrimSpace(event.VenueName)
	event.Address = strings.TrimSpace(event.Address)
	event.City = strings.TrimSpace(event.City)
	event.Description = sanitizeRichText(event.Description)
	event.Terms = sanitizeRichText(event.Terms)
	event.ContactName = strings.TrimSpace(event.ContactName)
	event.ContactEmail = emails.Normalize(event.ContactEmail)
	event.ContactPhone = strings.TrimSpace(event.ContactPhone)
	if event.CategoryId != nil && strings.TrimSpace(*event.CategoryId) == "" {
		event.CategoryId = nil
	}
	if event.GalleryUrls == nil {
		event.GalleryUrls = []string{}
	}
	if event.Latitude == nil || event.Longitude == nil {
		event.Latitude, event.Longitude = nil, nil
	}
	for index := range event.Tickets {
		event.Tickets[index].Name = strings.TrimSpace(event.Tickets[index].Name)
		event.Tickets[index].Description = strings.TrimSpace(event.Tickets[index].Description)
	}
}

// validateShape checks the fields that were filled in. It applies to drafts
// too, so a draft can be incomplete but never malformed.
func (s *EventService) validateShape(ctx context.Context, event *models.Event) map[string]string {
	fields := map[string]string{}
	set := func(field, message string) {
		if _, exists := fields[field]; !exists {
			fields[field] = message
		}
	}

	if event.Name == "" {
		set("name", "Event name is required.")
	} else if utf8.RuneCountInString(event.Name) > 150 {
		set("name", "Event name must be 150 characters or fewer.")
	}

	if event.CategoryId != nil {
		exists, err := s.categoryRepository.Exists(ctx, *event.CategoryId)
		if err != nil || !exists {
			set("categoryId", "Choose a category from the list.")
		}
	}

	prefix := imagePrefix(event.OrganizationId)
	if event.BannerUrl != "" && !s.storage.HasKeyPrefix(event.BannerUrl, prefix+"banner/") {
		set("bannerUrl", "Upload the banner again.")
	}
	if event.CoverUrl != "" && !s.storage.HasKeyPrefix(event.CoverUrl, prefix+"cover/") {
		set("coverUrl", "Upload the cover again.")
	}
	if len(event.GalleryUrls) > maxGalleryImages {
		set("galleryUrls", fmt.Sprintf("Add up to %d gallery photos.", maxGalleryImages))
	}
	for _, galleryUrl := range event.GalleryUrls {
		if !s.storage.HasKeyPrefix(galleryUrl, prefix+"gallery/") {
			set("galleryUrls", "Upload the gallery photos again.")
		}
	}

	if event.StartsAt != nil && event.EndsAt != nil && event.EndsAt.Before(*event.StartsAt) {
		set("endsAt", "The event must end after it starts.")
	}

	if event.Latitude != nil && (*event.Latitude < -90 || *event.Latitude > 90 || *event.Longitude < -180 || *event.Longitude > 180) {
		set("location", "Place the pin on the map again.")
	}
	if utf8.RuneCountInString(event.VenueName) > 150 {
		set("venueName", "Venue name must be 150 characters or fewer.")
	}
	if utf8.RuneCountInString(event.Address) > 500 {
		set("address", "Address must be 500 characters or fewer.")
	}
	if utf8.RuneCountInString(event.City) > 100 {
		set("city", "City must be 100 characters or fewer.")
	}

	for field, label := range map[string]string{"description": "Description", "terms": "Terms and agreement"} {
		value := event.Description
		if field == "terms" {
			value = event.Terms
		}
		if richTextLength(value) > maxRichTextLength {
			set(field, fmt.Sprintf("%s must be %d characters or fewer.", label, maxRichTextLength))
		} else if len(value) > maxRichTextBytes {
			set(field, fmt.Sprintf("%s has too much formatting. Simplify it and try again.", label))
		}
	}

	if utf8.RuneCountInString(event.ContactName) > 100 {
		set("contactName", "Contact name must be 100 characters or fewer.")
	}
	if event.ContactEmail != "" {
		if address, err := mail.ParseAddress(event.ContactEmail); err != nil || address.Address != event.ContactEmail || !strings.Contains(event.ContactEmail[strings.LastIndex(event.ContactEmail, "@"):], ".") {
			set("contactEmail", "Enter a valid email address.")
		}
	}
	if event.ContactPhone != "" && !phonePattern.MatchString(event.ContactPhone) {
		set("contactPhone", "Enter a valid phone number.")
	}

	switch event.ServiceFeeBearer {
	case "", models.ServiceFeeOrganizer, models.ServiceFeeBuyer:
	default:
		set("serviceFeeBearer", "Choose who pays the service fee.")
	}

	for field, message := range validateTickets(event) {
		set(field, message)
	}
	return fields
}

// validateTickets checks the shape of each ticket type (applies to drafts too).
func validateTickets(event *models.Event) map[string]string {
	fields := map[string]string{}
	set := func(field, message string) {
		if _, exists := fields[field]; !exists {
			fields[field] = message
		}
	}

	if len(event.Tickets) > maxTicketTypes {
		set("tickets", fmt.Sprintf("Add up to %d ticket types.", maxTicketTypes))
	}
	eventEnd := event.EndsAt
	if eventEnd == nil {
		eventEnd = event.StartsAt
	}
	for index := range event.Tickets {
		ticket := &event.Tickets[index]
		key := func(field string) string { return fmt.Sprintf("tickets.%d.%s", index, field) }

		switch ticket.Kind {
		case models.TicketPaid, models.TicketStreaming:
			if ticket.Price < 1 || ticket.Price > maxTicketPrice {
				set(key("price"), "Enter a price above zero.")
			}
		case models.TicketFree:
		default:
			set(key("kind"), "Choose a paid, free or streaming ticket.")
		}
		if ticket.Name == "" {
			set(key("name"), "Ticket name is required.")
		} else if utf8.RuneCountInString(ticket.Name) > 100 {
			set(key("name"), "Ticket name must be 100 characters or fewer.")
		}
		if utf8.RuneCountInString(ticket.Description) > 500 {
			set(key("description"), "Ticket description must be 500 characters or fewer.")
		}
		if ticket.Quantity < 1 || ticket.Quantity > maxTicketQuantity {
			set(key("quantity"), fmt.Sprintf("Quantity must be between 1 and %d.", maxTicketQuantity))
		}
		if ticket.SalesStartAt != nil && ticket.SalesEndAt != nil && ticket.SalesEndAt.Before(*ticket.SalesStartAt) {
			set(key("salesEndAt"), "Sales must end after they start.")
		}
		if ticket.SalesEndAt != nil && eventEnd != nil && ticket.SalesEndAt.After(*eventEnd) {
			set(key("salesEndAt"), "Sales must end before the event ends.")
		}
	}
	return fields
}

// validateComplete adds the rules only a published event must satisfy.
func validateComplete(event *models.Event, fields map[string]string) {
	set := func(field, message string) {
		if _, exists := fields[field]; !exists {
			fields[field] = message
		}
	}

	if event.CategoryId == nil {
		set("categoryId", "Choose a category.")
	}
	if event.BannerUrl == "" {
		set("bannerUrl", "Upload a banner.")
	}
	if event.StartsAt == nil {
		set("startsAt", "Choose when the event starts.")
	} else if !event.StartsAt.After(time.Now()) {
		set("startsAt", "The event must start in the future.")
	}
	if event.EndsAt == nil {
		set("endsAt", "Choose when the event ends.")
	}
	if needsVenue(event) {
		if event.Latitude == nil {
			set("location", "Place the venue pin on the map.")
		}
		if event.VenueName == "" {
			set("venueName", "Venue name is required.")
		}
		if event.Address == "" {
			set("address", "Address is required.")
		}
		if event.City == "" {
			set("city", "City is required.")
		}
	}
	if event.Description == "" {
		set("description", "Describe your event.")
	}
	if event.ContactName == "" {
		set("contactName", "Contact name is required.")
	}
	if event.ContactEmail == "" {
		set("contactEmail", "Contact email is required.")
	}
	if event.ContactPhone == "" {
		set("contactPhone", "Contact phone number is required.")
	}
	if len(event.Tickets) == 0 {
		set("tickets", "Add at least one ticket.")
	}
	for index := range event.Tickets {
		if event.Tickets[index].SalesStartAt == nil {
			set(fmt.Sprintf("tickets.%d.salesStartAt", index), "Choose when ticket sales start.")
		}
		if event.Tickets[index].SalesEndAt == nil {
			set(fmt.Sprintf("tickets.%d.salesEndAt", index), "Choose when ticket sales end.")
		}
	}
}

func imagePrefix(organizationId string) string {
	return "events/" + organizationId + "/"
}

// needsVenue reports whether any ticket grants venue entry; an event sold
// only as STREAMING tickets has no venue.
func needsVenue(event *models.Event) bool {
	for index := range event.Tickets {
		if event.Tickets[index].Kind != models.TicketStreaming {
			return true
		}
	}
	return false
}
