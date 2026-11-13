package handlers

import (
	"context"

	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/HappyNaCl/new-debyut/internal/users/services"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrganizerApplicationGrpcHandler struct {
	applicationSvc *services.OrganizerApplicationService

	users.UnimplementedOrganizerApplicationServiceServer
}

func NewOrganizerApplicationGrpcHandler(grpc *grpc.Server, service *services.OrganizerApplicationService) {
	handler := &OrganizerApplicationGrpcHandler{
		applicationSvc: service,
	}

	users.RegisterOrganizerApplicationServiceServer(grpc, handler)
}

func (h *OrganizerApplicationGrpcHandler) RequestEmailOtp(ctx context.Context, req *users.RequestEmailOtpRequest) (*emptypb.Empty, error) {
	if err := h.applicationSvc.RequestEmailOtp(ctx, req.GetUserId(), req.GetEmail()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *OrganizerApplicationGrpcHandler) VerifyEmailOtp(ctx context.Context, req *users.VerifyEmailOtpRequest) (*users.VerifyEmailOtpResponse, error) {
	verificationToken, err := h.applicationSvc.VerifyEmailOtp(ctx, req.GetUserId(), req.GetEmail(), req.GetOtp())
	if err != nil {
		return nil, err
	}
	return &users.VerifyEmailOtpResponse{VerificationToken: verificationToken}, nil
}

func (h *OrganizerApplicationGrpcHandler) SubmitApplication(ctx context.Context, req *users.SubmitApplicationRequest) (*users.OrganizerApplication, error) {
	application, err := h.applicationSvc.Submit(ctx, req.GetUserId(), req.GetVerificationToken(), req.GetOrganizationName(), req.GetDescription(), req.GetAvatarContent())
	if err != nil {
		return nil, err
	}
	return toApplicationResponse(application), nil
}

func (h *OrganizerApplicationGrpcHandler) ListMyApplications(ctx context.Context, req *users.ListMyApplicationsRequest) (*users.ListApplicationsResponse, error) {
	applications, err := h.applicationSvc.ListMine(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	return toListResponse(applications, int64(len(applications))), nil
}

func (h *OrganizerApplicationGrpcHandler) ListApplications(ctx context.Context, req *users.ListApplicationsRequest) (*users.ListApplicationsResponse, error) {
	var status models.ApplicationStatus
	if req.GetStatus() != users.ApplicationStatus_APPLICATION_STATUS_UNSPECIFIED {
		status = models.ApplicationStatus(req.GetStatus().String())
	}

	applications, total, err := h.applicationSvc.List(ctx, status, int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}
	return toListResponse(applications, total), nil
}

func (h *OrganizerApplicationGrpcHandler) ApproveApplication(ctx context.Context, req *users.ReviewApplicationRequest) (*users.OrganizerApplication, error) {
	application, err := h.applicationSvc.Approve(ctx, req.GetApplicationId(), req.GetReviewerId())
	if err != nil {
		return nil, err
	}
	return toApplicationResponse(application), nil
}

func (h *OrganizerApplicationGrpcHandler) DeclineApplication(ctx context.Context, req *users.ReviewApplicationRequest) (*users.OrganizerApplication, error) {
	application, err := h.applicationSvc.Decline(ctx, req.GetApplicationId(), req.GetReviewerId())
	if err != nil {
		return nil, err
	}
	return toApplicationResponse(application), nil
}

func toListResponse(applications []models.OrganizerApplication, total int64) *users.ListApplicationsResponse {
	response := &users.ListApplicationsResponse{
		Applications: make([]*users.OrganizerApplication, 0, len(applications)),
		Total:        total,
	}
	for index := range applications {
		response.Applications = append(response.Applications, toApplicationResponse(&applications[index]))
	}
	return response
}

func toApplicationResponse(application *models.OrganizerApplication) *users.OrganizerApplication {
	response := &users.OrganizerApplication{
		Id:                application.Id,
		UserId:            application.UserId,
		OrganizationName:  application.OrganizationName,
		OrganizationEmail: application.OrganizationEmail,
		Description:       application.Description,
		AvatarUrl:         application.AvatarUrl,
		Status:            users.ApplicationStatus(users.ApplicationStatus_value[string(application.Status)]),
		CreatedAt:         timestamppb.New(application.CreatedAt),
	}
	if application.ReviewedBy != nil {
		response.ReviewedBy = *application.ReviewedBy
	}
	if application.OrganizationId != nil {
		response.OrganizationId = *application.OrganizationId
	}
	if application.ReviewedAt != nil {
		response.ReviewedAt = timestamppb.New(*application.ReviewedAt)
	}
	return response
}
