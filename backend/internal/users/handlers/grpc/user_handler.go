package handlers

import (
	"context"

	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/HappyNaCl/new-debyut/internal/users/services"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserGrpcHandler struct {
	userSvc *services.UserService

	users.UnimplementedUserServiceServer
}

func NewUserGrpcHandler(grpc *grpc.Server, service *services.UserService) {
	handler := &UserGrpcHandler{
		userSvc: service,
	}

	users.RegisterUserServiceServer(grpc, handler)
}

func (h *UserGrpcHandler) GetMe(ctx context.Context, req *users.GetMeRequest) (*users.GetMeResponse, error) {
	user, err := h.userSvc.GetMe(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return toMeResponse(user), nil
}

func (h *UserGrpcHandler) UpdateMe(ctx context.Context, req *users.UpdateMeRequest) (*users.GetMeResponse, error) {
	user, err := h.userSvc.UpdateMe(ctx, req.GetId(), services.ProfileUpdate{
		FirstName:           req.GetFirstName(),
		Surname:             req.GetSurname(),
		DateOfBirth:         req.GetDateOfBirth().AsTime(),
		PhoneNumber:         req.GetPhoneNumber(),
		Gender:              models.Gender(req.GetGender()),
		SubscribeNewsletter: req.GetSubscribeNewsletter(),
	})
	if err != nil {
		return nil, err
	}

	return toMeResponse(user), nil
}

func (h *UserGrpcHandler) UpdateAvatar(ctx context.Context, req *users.UpdateAvatarRequest) (*users.GetMeResponse, error) {
	user, err := h.userSvc.UpdateAvatar(ctx, req.GetId(), req.GetContent())
	if err != nil {
		return nil, err
	}

	return toMeResponse(user), nil
}

func toMeResponse(user *models.User) *users.GetMeResponse {
	return &users.GetMeResponse{
		Id:                  user.Id,
		FirstName:           user.FirstName,
		Surname:             user.Surname,
		DateOfBirth:         timestamppb.New(user.DateOfBirth),
		PhoneNumber:         user.PhoneNumber,
		Gender:              string(user.Gender),
		SubscribeNewsletter: user.SubscribeNewsletter,
		AvatarUrl:           user.AvatarUrl,
	}
}
