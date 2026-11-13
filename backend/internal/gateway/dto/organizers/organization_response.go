package dto

import (
	"strings"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
)

type OrganizationResponse struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Description string    `json:"description"`
	AvatarUrl   string    `json:"avatarUrl"`
	MyRole      string    `json:"myRole,omitempty"`
	MemberCount int64     `json:"memberCount,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type OrganizationPageResponse struct {
	Organizations []*OrganizationResponse `json:"organizations"`
	Total         int64                   `json:"total"`
}

type InvitationResponse struct {
	Id               string    `json:"id"`
	OrganizationId   string    `json:"organizationId"`
	OrganizationName string    `json:"organizationName"`
	Email            string    `json:"email"`
	Role             string    `json:"role"`
	InvitedBy        string    `json:"invitedBy"`
	ExpiresAt        time.Time `json:"expiresAt"`
	CreatedAt        time.Time `json:"createdAt"`
}

type MemberResponse struct {
	UserId    string    `json:"userId"`
	FirstName string    `json:"firstName"`
	Surname   string    `json:"surname"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	JoinedAt  time.Time `json:"joinedAt"`
}

func NewOrganizationResponse(res *users.Organization) *OrganizationResponse {
	return &OrganizationResponse{
		Id:          res.GetId(),
		Name:        res.GetName(),
		Email:       res.GetEmail(),
		Description: res.GetDescription(),
		AvatarUrl:   res.GetAvatarUrl(),
		MemberCount: res.GetMemberCount(),
		MyRole:      roleName(res.GetMyRole()),
		CreatedAt:   res.GetCreatedAt().AsTime(),
	}
}

func NewOrganizationListResponse(res *users.ListOrganizationsResponse) []*OrganizationResponse {
	response := make([]*OrganizationResponse, 0, len(res.GetOrganizations()))
	for _, organization := range res.GetOrganizations() {
		response = append(response, NewOrganizationResponse(organization))
	}
	return response
}

func NewMemberResponse(res *users.OrganizationMember) *MemberResponse {
	return &MemberResponse{
		UserId:    res.GetUserId(),
		FirstName: res.GetFirstName(),
		Surname:   res.GetSurname(),
		Role:      res.GetRole().String(),
		Status:    strings.TrimPrefix(res.GetStatus().String(), "MEMBER_"),
		JoinedAt:  res.GetJoinedAt().AsTime(),
	}
}

func NewMemberListResponse(res *users.ListMembersResponse) []*MemberResponse {
	response := make([]*MemberResponse, 0, len(res.GetMembers()))
	for _, member := range res.GetMembers() {
		response = append(response, NewMemberResponse(member))
	}
	return response
}

func NewOrganizationPageResponse(res *users.ListOrganizationsResponse) *OrganizationPageResponse {
	return &OrganizationPageResponse{
		Organizations: NewOrganizationListResponse(res),
		Total:         res.GetTotal(),
	}
}

func NewInvitationResponse(res *users.Invitation) *InvitationResponse {
	return &InvitationResponse{
		Id:               res.GetId(),
		OrganizationId:   res.GetOrganizationId(),
		OrganizationName: res.GetOrganizationName(),
		Email:            res.GetEmail(),
		Role:             res.GetRole().String(),
		InvitedBy:        res.GetInvitedBy(),
		ExpiresAt:        res.GetExpiresAt().AsTime(),
		CreatedAt:        res.GetCreatedAt().AsTime(),
	}
}

func NewInvitationListResponse(res *users.ListInvitationsResponse) []*InvitationResponse {
	response := make([]*InvitationResponse, 0, len(res.GetInvitations()))
	for _, invitation := range res.GetInvitations() {
		response = append(response, NewInvitationResponse(invitation))
	}
	return response
}

func roleName(role users.MemberRole) string {
	if role == users.MemberRole_MEMBER_ROLE_UNSPECIFIED {
		return ""
	}
	return role.String()
}
