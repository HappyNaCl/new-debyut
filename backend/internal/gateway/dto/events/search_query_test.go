package dto

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func bindSearch(t *testing.T, rawQuery string) (SearchEventsQuery, error) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/events/search?"+rawQuery, nil)
	var query SearchEventsQuery
	err := c.ShouldBindQuery(&query)
	return query, err
}

func TestSearchEventsQueryBinding(t *testing.T) {
	query, err := bindSearch(t, "q=jazz&from=2026-11-14T10:00:00.000Z&to=2026-11-15T23:59:59.999Z&free=true&lat=-6.2&lng=106.8&radiusKm=25&sort=SOONEST&format=VENUE")
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	req := query.ToRequest()
	if req.GetFrom() == nil || req.GetTo() == nil {
		t.Fatalf("from/to not converted: %+v", req)
	}
	if !req.GetFree() || req.GetLatitude() != -6.2 || req.GetRadiusKm() != 25 || req.GetSort() != "SOONEST" {
		t.Errorf("unexpected request: %+v", req)
	}

	for _, bad := range []string{"from=yesterday", "sort=RANDOM", "radiusKm=0", "pageSize=100", "format=BOTH"} {
		if _, err := bindSearch(t, bad); err == nil {
			t.Errorf("%s: expected a validation error", bad)
		}
	}
}
