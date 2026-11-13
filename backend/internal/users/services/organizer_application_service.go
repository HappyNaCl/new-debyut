package services

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/common/emails"
	"github.com/HappyNaCl/new-debyut/internal/common/otp"
	"github.com/HappyNaCl/new-debyut/internal/common/rabbitmq"
	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	"github.com/HappyNaCl/new-debyut/internal/common/storage"
	"github.com/HappyNaCl/new-debyut/internal/common/tokens"
	topics "github.com/HappyNaCl/new-debyut/internal/constants"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/HappyNaCl/new-debyut/internal/users/repositories"
)

const (
	defaultPageSize = 20
	maximumPageSize = 100
)

type OrganizerApplicationService struct {
	applicationRepository  *repositories.OrganizerApplicationRepository
	organizationRepository *repositories.OrganizationRepository
	emailOtpCache          *otp.Cache
	emailTokenCache        *repositories.OrganizationEmailTokenCache
	publisher              *rabbitmq.Publisher
	storage                *storage.Client
}

func NewOrganizerApplicationService(
	applicationRepository *repositories.OrganizerApplicationRepository,
	organizationRepository *repositories.OrganizationRepository,
	emailOtpCache *otp.Cache,
	emailTokenCache *repositories.OrganizationEmailTokenCache,
	publisher *rabbitmq.Publisher,
	storage *storage.Client,
) *OrganizerApplicationService {
	return &OrganizerApplicationService{
		applicationRepository:  applicationRepository,
		organizationRepository: organizationRepository,
		emailOtpCache:          emailOtpCache,
		emailTokenCache:        emailTokenCache,
		publisher:              publisher,
		storage:                storage,
	}
}

func (s *OrganizerApplicationService) RequestEmailOtp(ctx context.Context, userId, email string) error {
	email = emails.Normalize(email)
	key := emailOtpKey(userId, email)

	emailInUse, err := s.organizationRepository.EmailInUse(ctx, email)
	if err != nil {
		return err
	}
	if emailInUse {
		return apperrors.ErrOrganizationEmailTaken
	}

	onCooldown, err := s.emailOtpCache.OnCooldown(ctx, key)
	if err != nil {
		return err
	}
	if onCooldown {
		return apperrors.ErrOtpResendCooldown
	}

	code := otp.Generate()
	if err := s.emailOtpCache.Put(ctx, key, code); err != nil {
		return err
	}

	return s.publish(ctx, topics.OTP_REQUESTED_TOPIC, events.OtpRequestedEvent{
		Email:   email,
		Otp:     code,
		Purpose: events.OtpPurposeOrganizationEmail,
	})
}

func (s *OrganizerApplicationService) VerifyEmailOtp(ctx context.Context, userId, email, code string) (string, error) {
	email = emails.Normalize(email)

	valid, err := s.emailOtpCache.Verify(ctx, emailOtpKey(userId, email), code)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", apperrors.ErrInvalidOrExpiredOtp
	}

	verificationToken, err := tokens.GetRefreshToken()
	if err != nil {
		return "", err
	}
	if err := s.emailTokenCache.Put(ctx, verificationToken, repositories.VerifiedOrganizationEmail{UserId: userId, Email: email}); err != nil {
		return "", err
	}
	return verificationToken, nil
}

func (s *OrganizerApplicationService) Submit(ctx context.Context, userId, verificationToken, organizationName, description string, avatarContent []byte) (*models.OrganizerApplication, error) {
	verified, err := s.emailTokenCache.Get(ctx, verificationToken)
	if err != nil || verified.UserId != userId {
		return nil, apperrors.ErrOrganizationEmailNotVerified
	}

	pending, err := s.applicationRepository.HasPending(ctx, userId)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, apperrors.ErrApplicationAlreadyPending
	}

	organizationName = strings.TrimSpace(organizationName)
	nameInUse, err := s.organizationRepository.NameInUse(ctx, organizationName, "")
	if err != nil {
		return nil, err
	}
	if nameInUse {
		return nil, apperrors.ErrOrganizationNameTaken
	}
	emailInUse, err := s.organizationRepository.EmailInUse(ctx, verified.Email)
	if err != nil {
		return nil, err
	}
	if emailInUse {
		return nil, apperrors.ErrOrganizationEmailTaken
	}

	applicationId := snowflake.Generate()
	avatarUrl := ""
	if len(avatarContent) > 0 {
		avatarUrl, err = uploadAvatar(ctx, s.storage, "applications", applicationId, avatarContent)
		if err != nil {
			return nil, err
		}
	}

	application := &models.OrganizerApplication{
		Id:                applicationId,
		UserId:            userId,
		OrganizationName:  organizationName,
		OrganizationEmail: verified.Email,
		Description:       description,
		AvatarUrl:         avatarUrl,
		Status:            models.ApplicationStatusPending,
	}
	if err := s.applicationRepository.Create(ctx, application); err != nil {
		deleteAvatar(ctx, s.storage, avatarUrl)
		return nil, err
	}

	_ = s.emailTokenCache.Delete(ctx, verificationToken)
	return application, nil
}

func (s *OrganizerApplicationService) ListMine(ctx context.Context, userId string) ([]models.OrganizerApplication, error) {
	return s.applicationRepository.ListByUser(ctx, userId)
}

func (s *OrganizerApplicationService) List(ctx context.Context, status models.ApplicationStatus, page, pageSize int) ([]models.OrganizerApplication, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	return s.applicationRepository.List(ctx, status, page, pageSize)
}

func (s *OrganizerApplicationService) Approve(ctx context.Context, id, reviewerId string) (*models.OrganizerApplication, error) {
	application, err := s.applicationRepository.Approve(ctx, id, reviewerId, &models.Organization{Id: snowflake.Generate()})
	if err != nil {
		return nil, err
	}
	if err := publishOrganizerStatus(ctx, s.publisher, application.UserId, true); err != nil {
		return nil, err
	}
	return application, nil
}

func (s *OrganizerApplicationService) Decline(ctx context.Context, id, reviewerId string) (*models.OrganizerApplication, error) {
	return s.applicationRepository.Decline(ctx, id, reviewerId)
}

func (s *OrganizerApplicationService) publish(ctx context.Context, topic string, event any) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.publisher.Publish(ctx, topic, eventBytes)
}

func publishOrganizerStatus(ctx context.Context, publisher *rabbitmq.Publisher, userId string, isOrganizer bool) error {
	eventBytes, err := json.Marshal(events.OrganizerStatusChangedEvent{
		UserId:      userId,
		IsOrganizer: isOrganizer,
	})
	if err != nil {
		return err
	}
	return publisher.Publish(ctx, topics.ORGANIZER_STATUS_CHANGED_TOPIC, eventBytes)
}

func emailOtpKey(userId, email string) string {
	return userId + ":" + email
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maximumPageSize {
		pageSize = maximumPageSize
	}
	return page, pageSize
}
