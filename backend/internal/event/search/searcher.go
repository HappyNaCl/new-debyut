package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/common/meilisearch"
	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
)

type Sort string

const (
	SortRelevance Sort = "RELEVANCE"
	SortSoonest   Sort = "SOONEST"
	SortNewest    Sort = "NEWEST"
	SortPriceLow  Sort = "PRICE_LOW"
)

type Format string

const (
	FormatAny    Format = ""
	FormatVenue  Format = "VENUE"
	FormatOnline Format = "ONLINE"
)

// Query is a public event search. Zero values mean "no filter".
type Query struct {
	Text         string
	CategorySlug string
	City         string
	// Events still running at or after From, and starting at or before To.
	From   *time.Time
	To     *time.Time
	Format Format
	Free   bool
	// Venues within RadiusKm of (Latitude, Longitude), when Near is set.
	Near      bool
	Latitude  float64
	Longitude float64
	RadiusKm  float64
	Sort      Sort
	Page      int
	PageSize  int
}

type Result struct {
	Ids   []string
	Total int64
	// Matching events per city and per category slug, ignoring pagination.
	Cities     map[string]int
	Categories map[string]int
}

type Searcher struct {
	client *meilisearch.Client
}

func NewSearcher(client *meilisearch.Client) *Searcher {
	return &Searcher{client: client}
}

func (s *Searcher) Search(ctx context.Context, query Query) (*Result, error) {
	filters := []string{"endsAt >= " + strconv.FormatInt(time.Now().Unix(), 10)}
	if query.CategorySlug != "" {
		filters = append(filters, "categorySlug = "+meilisearch.Quote(query.CategorySlug))
	}
	if query.City != "" {
		filters = append(filters, "city = "+meilisearch.Quote(query.City))
	}
	if query.From != nil {
		filters = append(filters, "endsAt >= "+strconv.FormatInt(query.From.Unix(), 10))
	}
	if query.To != nil {
		filters = append(filters, "startsAt <= "+strconv.FormatInt(query.To.Unix(), 10))
	}
	switch query.Format {
	case FormatVenue:
		filters = append(filters, "hasVenue = true")
	case FormatOnline:
		filters = append(filters, "hasStreaming = true")
	}
	if query.Free {
		filters = append(filters, "isFree = true")
	}
	if query.Near {
		filters = append(filters, fmt.Sprintf("_geoRadius(%s, %s, %d)",
			strconv.FormatFloat(query.Latitude, 'f', -1, 64),
			strconv.FormatFloat(query.Longitude, 'f', -1, 64),
			int64(query.RadiusKm*1000)))
	}

	var sort []string
	switch query.Sort {
	case SortSoonest:
		sort = []string{"startsAt:asc"}
	case SortNewest:
		sort = []string{"publishedAt:desc"}
	case SortPriceLow:
		sort = []string{"startingPrice:asc", "startsAt:asc"}
	default:
		// Relevance only means something with search text; without it,
		// show what's coming up first.
		if strings.TrimSpace(query.Text) == "" {
			sort = []string{"startsAt:asc"}
		}
	}

	res, err := s.client.Search(ctx, IndexName, meilisearch.SearchRequest{
		Query:                query.Text,
		Filter:               filters,
		Sort:                 sort,
		Page:                 query.Page,
		HitsPerPage:          query.PageSize,
		Facets:               []string{"city", "categorySlug"},
		AttributesToRetrieve: []string{"id"},
	})
	if err != nil {
		return nil, err
	}

	result := &Result{
		Ids:        make([]string, 0, len(res.Hits)),
		Total:      res.TotalHits,
		Cities:     res.FacetDistribution["city"],
		Categories: res.FacetDistribution["categorySlug"],
	}
	for _, raw := range res.Hits {
		var hit struct {
			Id string `json:"id"`
		}
		if err := json.Unmarshal(raw, &hit); err != nil {
			return nil, err
		}
		result.Ids = append(result.Ids, hit.Id)
	}
	return result, nil
}

// OrganizationDirectory looks organizer names up in the users service.
type OrganizationDirectory struct {
	client users.OrganizationServiceClient
}

func NewOrganizationDirectory(client users.OrganizationServiceClient) *OrganizationDirectory {
	return &OrganizationDirectory{client: client}
}

func (d *OrganizationDirectory) Names(ctx context.Context, organizationIds []string) (map[string]string, error) {
	res, err := d.client.ListOrganizationProfiles(ctx, &users.ListOrganizationProfilesRequest{OrganizationIds: organizationIds})
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(res.GetOrganizations()))
	for _, profile := range res.GetOrganizations() {
		names[profile.GetId()] = profile.GetName()
	}
	return names, nil
}
