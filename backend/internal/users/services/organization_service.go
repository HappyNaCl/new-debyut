package services

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/rabbitmq"
	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	"github.com/HappyNaCl/new-debyut/internal/common/storage"
	topics "github.com/HappyNaCl/new-debyut/internal/constants"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/HappyNaCl/new-debyut/internal/users/repositories"
	"go.uber.org/zap"
)

const invitationLifetime = 7 * 24 * time.Hour

type OrganizationService struct {
	organizationRepository *repositories.OrganizationRepository
	invitationRepository   *repositories.InvitationRepository
	userRepository         *repositories.UserRepository
	publisher              *rabbitmq.Publisher
	storage                *storage.Client
}

func NewOrganizationService(
	organizationRepository *repositories.OrganizationRepository,
	invitationRepository *repositories.InvitationRepository,
	userRepository *repositories.UserRepository,
	publisher *rabbitmq.Publisher,
	storage *storage.Client,
) *OrganizationService {
	return &OrganizationService{
		organizationRepository: organizationRepository,
		invitationRepository:   invitationRepository,
		userRepository:         userRepository,
		publisher:              publisher,
		storage:                storage,
	}
}

func (s *OrganizationService) ListMine(ctx context.Context, userId string) ([]models.MembershipOrganization, error) {
	return s.organizationRepository.ListByMember(ctx, userId)
}

func (s *OrganizationService) List(ctx context.Context, search string, page, pageSize int) ([]models.OrganizationSummary, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	return s.organizationRepository.List(ctx, strings.TrimSpace(search), page, pageSize)
}

func (s *OrganizationService) Get(ctx context.Context, organizationId, actorId string) (*models.MembershipOrganization, error) {
	return s.organizationRepository.GetForMember(ctx, organizationId, actorId)
}

func (s *OrganizationService) Update(ctx context.Context, organizationId, actorId, name, description string) (*models.MembershipOrganization, error) {
	organization, err := s.requireOwner(ctx, organizationId, actorId)
	if err != nil {
		return nil, err
	}

	if err := s.organizationRepository.Update(ctx, organizationId, name, description); err != nil {
		return nil, err
	}

	organization.Name = name
	organization.Description = description
	return organization, nil
}

func (s *OrganizationService) UpdateAvatar(ctx context.Context, organizationId, actorId string, content []byte) (*models.MembershipOrganization, error) {
	organization, err := s.requireOwner(ctx, organizationId, actorId)
	if err != nil {
		return nil, err
	}

	url, err := uploadAvatar(ctx, s.storage, "organizations", organizationId, content)
	if err != nil {
		return nil, err
	}

	if err := s.organizationRepository.UpdateAvatar(ctx, organizationId, url); err != nil {
		deleteAvatar(ctx, s.storage, url)
		return nil, err
	}

	deleteAvatar(ctx, s.storage, organization.AvatarUrl)
	organization.AvatarUrl = url
	return organization, nil
}

func (s *OrganizationService) ListMembers(ctx context.Context, organizationId, actorId string) ([]models.MemberProfile, error) {
	if _, err := s.organizationRepository.GetForMember(ctx, organizationId, actorId); err != nil {
		return nil, err
	}
	return s.organizationRepository.ListMembers(ctx, organizationId)
}

func (s *OrganizationService) RemoveMember(ctx context.Context, organizationId, actorId, userId string) error {
	if actorId != userId {
		if _, err := s.requireOwner(ctx, organizationId, actorId); err != nil {
			return err
		}
	}

	if err := s.organizationRepository.RemoveMember(ctx, organizationId, userId); err != nil {
		return err
	}

	remaining, err := s.organizationRepository.CountMemberships(ctx, userId)
	if err != nil {
		return err
	}
	if remaining == 0 {
		return publishOrganizerStatus(ctx, s.publisher, userId, false)
	}
	return nil
}

func (s *OrganizationService) CreateInvitation(ctx context.Context, organizationId, actorId, email string, role models.MemberRole) (*models.OrganizationInvitation, error) {
	if role != models.MemberRoleOwner && role != models.MemberRoleStaff {
		return nil, apperrors.ErrInvalidMemberRole
	}
	organization, err := s.requireOwner(ctx, organizationId, actorId)
	if err != nil {
		return nil, err
	}

	invitation := &models.OrganizationInvitation{
		Id:             snowflake.Generate(),
		OrganizationId: organizationId,
		Email:          normalizeEmail(email),
		Role:           role,
		Status:         models.InvitationStatusPending,
		InvitedBy:      actorId,
		ExpiresAt:      time.Now().Add(invitationLifetime).UTC(),
	}
	if err := s.invitationRepository.Create(ctx, invitation); err != nil {
		return nil, err
	}
	invitation.Organization = organization.Organization

	inviterName := "A debyut organizer"
	if inviter, err := s.userRepository.GetById(ctx, actorId); err == nil {
		inviterName = strings.TrimSpace(inviter.FirstName + " " + inviter.Surname)
	}

	eventBytes, err := json.Marshal(events.OrganizationInvitationCreatedEvent{
		InvitationId:     invitation.Id,
		OrganizationName: organization.Name,
		Email:            invitation.Email,
		Role:             string(invitation.Role),
		InviterName:      inviterName,
		ExpiresAt:        invitation.ExpiresAt,
	})
	if err == nil {
		err = s.publisher.Publish(ctx, topics.ORGANIZATION_INVITATION_CREATED_TOPIC, eventBytes)
	}
	if err != nil {
		l.Logger.Warn("Failed to publish invitation email event", zap.String("invitationId", invitation.Id), zap.Error(err))
	}

	return invitation, nil
}

func (s *OrganizationService) ListInvitations(ctx context.Context, organizationId, actorId string) ([]models.OrganizationInvitation, error) {
	if _, err := s.requireOwner(ctx, organizationId, actorId); err != nil {
		return nil, err
	}
	return s.invitationRepository.ListPendingByOrganization(ctx, organizationId)
}

func (s *OrganizationService) RevokeInvitation(ctx context.Context, organizationId, invitationId, actorId string) error {
	if _, err := s.requireOwner(ctx, organizationId, actorId); err != nil {
		return err
	}
	return s.invitationRepository.Revoke(ctx, invitationId, organizationId)
}

func (s *OrganizationService) ListMyInvitations(ctx context.Context, email string) ([]models.OrganizationInvitation, error) {
	return s.invitationRepository.ListPendingByEmail(ctx, email)
}

func (s *OrganizationService) AcceptInvitation(ctx context.Context, invitationId, actorId, actorEmail string) (*models.MemberProfile, error) {
	invitation, err := s.invitationRepository.Accept(ctx, invitationId, actorId, actorEmail)
	if err != nil {
		return nil, err
	}
	if err := publishOrganizerStatus(ctx, s.publisher, actorId, true); err != nil {
		return nil, err
	}

	members, err := s.organizationRepository.ListMembers(ctx, invitation.OrganizationId)
	if err != nil {
		return nil, err
	}
	for index := range members {
		if members[index].UserId == actorId {
			return &members[index], nil
		}
	}
	return nil, apperrors.ErrMemberNotFound
}

func (s *OrganizationService) DeclineInvitation(ctx context.Context, invitationId, actorEmail string) error {
	return s.invitationRepository.Decline(ctx, invitationId, actorEmail)
}

func (s *OrganizationService) requireOwner(ctx context.Context, organizationId, actorId string) (*models.MembershipOrganization, error) {
	organization, err := s.organizationRepository.GetForMember(ctx, organizationId, actorId)
	if err != nil {
		return nil, err
	}
	if organization.Role != models.MemberRoleOwner {
		return nil, apperrors.ErrNotOrganizationOwner
	}
	return organization, nil
}
