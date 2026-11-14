// Package meilisearch is a small client for the parts of the Meilisearch HTTP
// API that debyut uses: index settings, document writes and search.
package meilisearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseUrl string
	apiKey  string
	http    *http.Client
}

// Connect builds a client from MEILISEARCH_URL and MEILISEARCH_MASTER_KEY and
// checks that the server answers.
func Connect(ctx context.Context) (*Client, error) {
	baseUrl := strings.TrimRight(os.Getenv("MEILISEARCH_URL"), "/")
	if baseUrl == "" {
		return nil, errors.New("MEILISEARCH_URL is not set")
	}
	client := &Client{
		baseUrl: baseUrl,
		apiKey:  os.Getenv("MEILISEARCH_MASTER_KEY"),
		http:    &http.Client{Timeout: 15 * time.Second},
	}
	if err := client.do(ctx, http.MethodGet, "/health", nil, nil); err != nil {
		return nil, fmt.Errorf("meilisearch health check: %w", err)
	}
	return client, nil
}

// Error is a non-2xx answer from Meilisearch.
type Error struct {
	Status  int
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("meilisearch %d %s: %s", e.Status, e.Code, e.Message)
}

func IsNotFound(err error) bool {
	var apiErr *Error
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseUrl+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		apiErr := &Error{Status: res.StatusCode}
		_ = json.NewDecoder(res.Body).Decode(apiErr)
		return apiErr
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, res.Body)
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// Task is an asynchronous write; Meilisearch applies writes in the order it
// accepted them.
type Task struct {
	Uid    int64  `json:"taskUid"`
	Status string `json:"status"`
}

type taskState struct {
	Status string `json:"status"`
	Error  *Error `json:"error"`
}

// Wait blocks until the task has been applied and reports whether it failed.
func (c *Client) Wait(ctx context.Context, task Task) error {
	delay := 25 * time.Millisecond
	for {
		var state taskState
		if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/tasks/%d", task.Uid), nil, &state); err != nil {
			return err
		}
		switch state.Status {
		case "succeeded":
			return nil
		case "failed", "canceled":
			if state.Error != nil {
				return state.Error
			}
			return fmt.Errorf("meilisearch task %d %s", task.Uid, state.Status)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay = min(delay*2, 500*time.Millisecond)
	}
}

func (c *Client) run(ctx context.Context, method, path string, body any) error {
	var task Task
	if err := c.do(ctx, method, path, body, &task); err != nil {
		return err
	}
	return c.Wait(ctx, task)
}

// EnsureIndex creates the index if it's missing and applies its settings.
// It reports whether the index was created, i.e. starts out empty.
func (c *Client) EnsureIndex(ctx context.Context, index, primaryKey string, settings any) (bool, error) {
	created := false
	err := c.do(ctx, http.MethodGet, "/indexes/"+url.PathEscape(index), nil, nil)
	if IsNotFound(err) {
		if err := c.run(ctx, http.MethodPost, "/indexes", map[string]string{"uid": index, "primaryKey": primaryKey}); err != nil {
			return false, err
		}
		created = true
	} else if err != nil {
		return false, err
	}
	return created, c.run(ctx, http.MethodPatch, "/indexes/"+url.PathEscape(index)+"/settings", settings)
}

// DocumentCount reports how many documents an index holds.
func (c *Client) DocumentCount(ctx context.Context, index string) (int64, error) {
	var stats struct {
		NumberOfDocuments int64 `json:"numberOfDocuments"`
	}
	err := c.do(ctx, http.MethodGet, "/indexes/"+url.PathEscape(index)+"/stats", nil, &stats)
	return stats.NumberOfDocuments, err
}

// UpsertDocuments replaces documents by primary key and waits until applied.
func (c *Client) UpsertDocuments(ctx context.Context, index string, documents any) error {
	return c.run(ctx, http.MethodPost, "/indexes/"+url.PathEscape(index)+"/documents", documents)
}

// DeleteDocuments removes documents by primary key and waits until applied.
func (c *Client) DeleteDocuments(ctx context.Context, index string, ids []string) error {
	return c.run(ctx, http.MethodPost, "/indexes/"+url.PathEscape(index)+"/documents/delete-batch", ids)
}

// DeleteByFilter removes every document matching a filter expression.
func (c *Client) DeleteByFilter(ctx context.Context, index, filter string) error {
	return c.run(ctx, http.MethodPost, "/indexes/"+url.PathEscape(index)+"/documents/delete", map[string]string{"filter": filter})
}

type SearchRequest struct {
	Query                string   `json:"q,omitempty"`
	Filter               []string `json:"filter,omitempty"`
	Sort                 []string `json:"sort,omitempty"`
	Page                 int      `json:"page,omitempty"`
	HitsPerPage          int      `json:"hitsPerPage,omitempty"`
	Facets               []string `json:"facets,omitempty"`
	AttributesToRetrieve []string `json:"attributesToRetrieve,omitempty"`
}

type SearchResponse struct {
	Hits              []json.RawMessage         `json:"hits"`
	TotalHits         int64                     `json:"totalHits"`
	FacetDistribution map[string]map[string]int `json:"facetDistribution"`
}

func (c *Client) Search(ctx context.Context, index string, req SearchRequest) (*SearchResponse, error) {
	var res SearchResponse
	if err := c.do(ctx, http.MethodPost, "/indexes/"+url.PathEscape(index)+"/search", req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Quote escapes a value for use inside a filter expression.
func Quote(value string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value) + `"`
}
