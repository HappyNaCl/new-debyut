package handlers

import (
	"context"
	"strings"

	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/HappyNaCl/new-debyut/internal/users/services"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrganizationGrpcHandler struct {
	organizationSvc *services.OrganizationService

	users.UnimplementedOrganizationServiceServer
}

func NewOrganizationGrpcHandler(grpc *grpc.Server, service *services.OrganizationService) {
	handler := &OrganizationGrpcHandler{
		organizationSvc: service,
	}

	users.RegisterOrganizationServiceServer(grpc, handler)
}

func (h *OrganizationGrpcHandler) ListMyOrganizations(ctx context.Context, req *users.ListMyOrganizationsRequest) (*users.ListOrganizationsResponse, error) {
	organizations, err := h.organizationSvc.ListMine(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	response := &users.ListOrganizationsResponse{
		Organizations: make([]*users.Organization, 0, len(organizations)),
		Total:         int64(len(organizations)),
	}
	for index := range organizations {
		organization := toOrganizationResponse(&organizations[index].Organization)
		organization.MyRole = toMemberRole(organizations[index].Role)
		response.Organizations = append(response.Organizations, organization)
	}
	return response, nil
}

func (h *OrganizationGrpcHandler) ListOrganizations(ctx context.Context, req *users.ListOrganizationsRequest) (*users.ListOrganizationsResponse, error) {
	organizations, total, err := h.organizationSvc.List(ctx, req.GetSearch(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, err
	}

	response := &users.ListOrganizationsResponse{
		Organizations: make([]*users.Organization, 0, len(organizations)),
		Total:         total,
	}
	for index := range organizations {
		organization := toOrganizationResponse(&organizations[index].Organization)
		organization.MemberCount = organizations[index].MemberCount
		response.Organizations = append(response.Organizations, organization)
	}
	return response, nil
}

func (h *OrganizationGrpcHandler) GetOrganizationProfile(ctx context.Context, req *users.OrganizationProfileRequest) (*users.OrganizationProfile, error) {
	organization, err := h.organizationSvc.GetProfile(ctx, req.GetOrganizationId())
	if err != nil {
		return nil, err
	}
	return &users.OrganizationProfile{Id: organization.Id, Name: organization.Name, AvatarUrl: organization.AvatarUrl}, nil
}

func (h *OrganizationGrpcHandler) ListOrganizationProfiles(ctx context.Context, req *users.ListOrganizationProfilesRequest) (*users.ListOrganizationProfilesResponse, error) {
	organizations, err := h.organizationSvc.ListProfiles(ctx, req.GetOrganizationIds())
	if err != nil {
		return nil, err
	}
	response := &users.ListOrganizationProfilesResponse{Organizations: make([]*users.OrganizationProfile, 0, len(organizations))}
	for index := range organizations {
		organization := &organizations[index]
		response.Organizations = append(response.Organizations, &users.OrganizationProfile{
			Id: organization.Id, Name: organization.Name, AvatarUrl: organization.AvatarUrl,
		})
	}
	return response, nil
}

func (h *OrganizationGrpcHandler) GetOrganization(ctx context.Context, req *users.OrganizationRequest) (*users.Organization, error) {
	organization, err := h.organizationSvc.Get(ctx, req.GetOrganizationId(), req.GetActorId())
	if err != nil {
		return nil, err
	}

	response := toOrganizationResponse(&organization.Organization)
	response.MyRole = toMemberRole(organization.Role)
	return response, nil
}

func (h *OrganizationGrpcHandler) UpdateOrganization(ctx context.Context, req *users.UpdateOrganizationRequest) (*users.Organization, error) {
	organization, err := h.organizationSvc.Update(ctx, req.GetOrganizationId(), req.GetActorId(), req.GetName(), req.GetDescription())
	if err != nil {
		return nil, err
	}

	response := toOrganizationResponse(&organization.Organization)
	response.MyRole = toMemberRole(organization.Role)
	return response, nil
}

func (h *OrganizationGrpcHandler) UpdateOrganizationAvatar(ctx context.Context, req *users.UpdateOrganizationAvatarRequest) (*users.Organization, error) {
	organization, err := h.organizationSvc.UpdateAvatar(ctx, req.GetOrganizationId(), req.GetActorId(), req.GetContent())
	if err != nil {
		return nil, err
	}

	response := toOrganizationResponse(&organization.Organization)
	response.MyRole = toMemberRole(organization.Role)
	return response, nil
}

func (h *OrganizationGrpcHandler) ListMembers(ctx context.Context, req *users.OrganizationRequest) (*users.ListMembersResponse, error) {
	members, err := h.organizationSvc.ListMembers(ctx, req.GetOrganizationId(), req.GetActorId())
	if err != nil {
		return nil, err
	}

	response := &users.ListMembersResponse{Members: make([]*users.OrganizationMember, 0, len(members))}
	for index := range members {
		response.Members = append(response.Members, toMemberResponse(&members[index]))
	}
	return response, nil
}

func (h *OrganizationGrpcHandler) RemoveMember(ctx context.Context, req *users.RemoveMemberRequest) (*emptypb.Empty, error) {
	if err := h.organizationSvc.RemoveMember(ctx, req.GetOrganizationId(), req.GetActorId(), req.GetUserId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *OrganizationGrpcHandler) SetMemberStatus(ctx context.Context, req *users.SetMemberStatusRequest) (*users.OrganizationMember, error) {
	status := models.MemberStatus(strings.TrimPrefix(req.GetStatus().String(), "MEMBER_"))
	member, err := h.organizationSvc.SetMemberStatus(ctx, req.GetOrganizationId(), req.GetActorId(), req.GetUserId(), status)
	if err != nil {
		return nil, err
	}
	return toMemberResponse(member), nil
}

func (h *OrganizationGrpcHandler) CreateInvitation(ctx context.Context, req *users.CreateInvitationRequest) (*users.Invitation, error) {
	invitation, err := h.organizationSvc.CreateInvitation(ctx, req.GetOrganizationId(), req.GetActorId(), req.GetEmail(), models.MemberRole(req.GetRole().String()))
	if err != nil {
		return nil, err
	}
	return toInvitationResponse(invitation), nil
}

func (h *OrganizationGrpcHandler) ListInvitations(ctx context.Context, req *users.OrganizationRequest) (*users.ListInvitationsResponse, error) {
	invitations, err := h.organizationSvc.ListInvitations(ctx, req.GetOrganizationId(), req.GetActorId())
	if err != nil {
		return nil, err
	}
	return toInvitationListResponse(invitations), nil
}

func (h *OrganizationGrpcHandler) RevokeInvitation(ctx context.Context, req *users.InvitationActionRequest) (*emptypb.Empty, error) {
	if err := h.organizationSvc.RevokeInvitation(ctx, req.GetOrganizationId(), req.GetInvitationId(), req.GetActorId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *OrganizationGrpcHandler) ListMyInvitations(ctx context.Context, req *users.ListMyInvitationsRequest) (*users.ListInvitationsResponse, error) {
	invitations, err := h.organizationSvc.ListMyInvitations(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	return toInvitationListResponse(invitations), nil
}

func (h *OrganizationGrpcHandler) AcceptInvitation(ctx context.Context, req *users.InvitationActionRequest) (*users.OrganizationMember, error) {
	member, err := h.organizationSvc.AcceptInvitation(ctx, req.GetInvitationId(), req.GetActorId(), req.GetActorEmail())
	if err != nil {
		return nil, err
	}
	return toMemberResponse(member), nil
}

func (h *OrganizationGrpcHandler) DeclineInvitation(ctx context.Context, req *users.InvitationActionRequest) (*emptypb.Empty, error) {
	if err := h.organizationSvc.DeclineInvitation(ctx, req.GetInvitationId(), req.GetActorEmail()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func toMemberRole(role models.MemberRole) users.MemberRole {
	return users.MemberRole(users.MemberRole_value[string(role)])
}

func toOrganizationResponse(organization *models.Organization) *users.Organization {
	return &users.Organization{
		Id:          organization.Id,
		Name:        organization.Name,
		Email:       organization.Email,
		Description: organization.Description,
		AvatarUrl:   organization.AvatarUrl,
		CreatedAt:   timestamppb.New(organization.CreatedAt),
	}
}

func toMemberResponse(member *models.MemberProfile) *users.OrganizationMember {
	return &users.OrganizationMember{
		OrganizationId: member.OrganizationId,
		UserId:         member.UserId,
		FirstName:      member.FirstName,
		Surname:        member.Surname,
		Role:           toMemberRole(member.Role),
		Status:         users.MemberStatus(users.MemberStatus_value["MEMBER_"+string(member.Status)]),
		JoinedAt:       timestamppb.New(member.CreatedAt),
	}
}

func toInvitationResponse(invitation *models.OrganizationInvitation) *users.Invitation {
	return &users.Invitation{
		Id:               invitation.Id,
		OrganizationId:   invitation.OrganizationId,
		OrganizationName: invitation.Organization.Name,
		Email:            invitation.Email,
		Role:             toMemberRole(invitation.Role),
		Status:           users.InvitationStatus(users.InvitationStatus_value["INVITATION_"+string(invitation.Status)]),
		InvitedBy:        invitation.InvitedBy,
		ExpiresAt:        timestamppb.New(invitation.ExpiresAt),
		CreatedAt:        timestamppb.New(invitation.CreatedAt),
	}
}

func toInvitationListResponse(invitations []models.OrganizationInvitation) *users.ListInvitationsResponse {
	response := &users.ListInvitationsResponse{Invitations: make([]*users.Invitation, 0, len(invitations))}
	for index := range invitations {
		response.Invitations = append(response.Invitations, toInvitationResponse(&invitations[index]))
	}
	return response
}
