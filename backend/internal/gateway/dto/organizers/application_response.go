package dto

import (
	"time"

	"github.com/HappyNaCl/new-debyut/internal/genproto/users"
)

type ApplicationResponse struct {
	Id                string     `json:"id"`
	UserId            string     `json:"userId"`
	OrganizationName  string     `json:"organizationName"`
	OrganizationEmail string     `json:"organizationEmail"`
	Description       string     `json:"description"`
	AvatarUrl         string     `json:"avatarUrl,omitempty"`
	Status            string     `json:"status"`
	ReviewedBy        string     `json:"reviewedBy,omitempty"`
	ReviewedAt        *time.Time `json:"reviewedAt,omitempty"`
	OrganizationId    string     `json:"organizationId,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type ApplicationListResponse struct {
	Applications []*ApplicationResponse `json:"applications"`
	Total        int64                  `json:"total"`
}

func NewApplicationResponse(res *users.OrganizerApplication) *ApplicationResponse {
	response := &ApplicationResponse{
		Id:                res.GetId(),
		UserId:            res.GetUserId(),
		OrganizationName:  res.GetOrganizationName(),
		OrganizationEmail: res.GetOrganizationEmail(),
		Description:       res.GetDescription(),
		AvatarUrl:         res.GetAvatarUrl(),
		Status:            res.GetStatus().String(),
		ReviewedBy:        res.GetReviewedBy(),
		OrganizationId:    res.GetOrganizationId(),
		CreatedAt:         res.GetCreatedAt().AsTime(),
	}
	if res.GetReviewedAt() != nil {
		reviewedAt := res.GetReviewedAt().AsTime()
		response.ReviewedAt = &reviewedAt
	}
	return response
}

func NewApplicationListResponse(res *users.ListApplicationsResponse) *ApplicationListResponse {
	response := &ApplicationListResponse{
		Applications: make([]*ApplicationResponse, 0, len(res.GetApplications())),
		Total:        res.GetTotal(),
	}
	for _, application := range res.GetApplications() {
		response.Applications = append(response.Applications, NewApplicationResponse(application))
	}
	return response
}
