package dto

import (
	"time"

	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
)

type CategoryResponse struct {
	Id        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewCategoryResponse(res *event.Category) *CategoryResponse {
	return &CategoryResponse{
		Id:        res.GetId(),
		Slug:      res.GetSlug(),
		Name:      res.GetName(),
		Icon:      res.GetIcon(),
		CreatedAt: res.GetCreatedAt().AsTime(),
		UpdatedAt: res.GetUpdatedAt().AsTime(),
	}
}

func NewCategoryListResponse(res *event.ListCategoriesResponse) []*CategoryResponse {
	response := make([]*CategoryResponse, 0, len(res.GetCategories()))
	for _, category := range res.GetCategories() {
		response = append(response, NewCategoryResponse(category))
	}
	return response
}
