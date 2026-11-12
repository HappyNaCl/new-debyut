package handlers

import (
	"context"

	"github.com/HappyNaCl/new-debyut/internal/event/models"
	"github.com/HappyNaCl/new-debyut/internal/event/services"
	"github.com/HappyNaCl/new-debyut/internal/genproto/event"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventGrpcHandler struct {
	categorySvc *services.CategoryService

	event.UnimplementedEventServiceServer
}

func NewEventGrpcHandler(grpc *grpc.Server, categoryService *services.CategoryService) {
	handler := &EventGrpcHandler{
		categorySvc: categoryService,
	}

	event.RegisterEventServiceServer(grpc, handler)
}

func (h *EventGrpcHandler) ListCategories(ctx context.Context, _ *emptypb.Empty) (*event.ListCategoriesResponse, error) {
	categories, err := h.categorySvc.List(ctx)
	if err != nil {
		return nil, err
	}

	response := &event.ListCategoriesResponse{Categories: make([]*event.Category, 0, len(categories))}
	for index := range categories {
		response.Categories = append(response.Categories, toCategoryResponse(&categories[index]))
	}
	return response, nil
}

func (h *EventGrpcHandler) CreateCategory(ctx context.Context, req *event.CreateCategoryRequest) (*event.Category, error) {
	category, err := h.categorySvc.Create(ctx, req.GetName(), req.GetSlug(), req.GetIcon())
	if err != nil {
		return nil, err
	}
	return toCategoryResponse(category), nil
}

func (h *EventGrpcHandler) UpdateCategory(ctx context.Context, req *event.UpdateCategoryRequest) (*event.Category, error) {
	category, err := h.categorySvc.Update(ctx, req.GetId(), req.GetName(), req.GetSlug(), req.GetIcon())
	if err != nil {
		return nil, err
	}
	return toCategoryResponse(category), nil
}

func (h *EventGrpcHandler) DeleteCategory(ctx context.Context, req *event.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if err := h.categorySvc.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func toCategoryResponse(category *models.Category) *event.Category {
	return &event.Category{
		Id:        category.Id,
		Slug:      category.Slug,
		Name:      category.Name,
		Icon:      category.Icon,
		CreatedAt: timestamppb.New(category.CreatedAt),
		UpdatedAt: timestamppb.New(category.UpdatedAt),
	}
}
