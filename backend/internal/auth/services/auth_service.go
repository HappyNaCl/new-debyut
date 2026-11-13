package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/auth/models"
	"github.com/HappyNaCl/new-debyut/internal/auth/repositories"
	"github.com/HappyNaCl/new-debyut/internal/common/emails"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/otp"
	"github.com/HappyNaCl/new-debyut/internal/common/rabbitmq"
	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	"github.com/HappyNaCl/new-debyut/internal/common/tokens"
	topics "github.com/HappyNaCl/new-debyut/internal/constants"
	"github.com/HappyNaCl/new-debyut/internal/events"
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
)

const minimumAge = 18

type AuthResult struct {
	Credential *models.UserCredential
	TokenPair  *TokenPair
}

type AuthService struct {
	registerOtpCache       *otp.Cache
	loginOtpCache          *otp.Cache
	registrationTokenCache *repositories.RegistrationTokenCache
	credentialRepository   *repositories.CredentialRepository
	tokenService           *TokenService
	publisher              *rabbitmq.Publisher
}

func NewAuthService(
	registerOtpCache *otp.Cache,
	loginOtpCache *otp.Cache,
	registrationTokenCache *repositories.RegistrationTokenCache,
	credentialRepo *repositories.CredentialRepository,
	tokenSvc *TokenService,
	publisher *rabbitmq.Publisher) *AuthService {
	return &AuthService{
		registerOtpCache:       registerOtpCache,
		loginOtpCache:          loginOtpCache,
		registrationTokenCache: registrationTokenCache,
		credentialRepository:   credentialRepo,
		tokenService:           tokenSvc,
		publisher:              publisher,
	}
}

func (a *AuthService) RequestRegisterOtp(ctx context.Context, email string) error {
	email = emails.Normalize(email)
	if _, err := a.credentialRepository.GetByEmail(ctx, email); err == nil {
		return apperrors.ErrEmailAlreadyExists
	}
	return a.sendOtp(ctx, a.registerOtpCache, email)
}

func (a *AuthService) VerifyRegisterOtp(ctx context.Context, email, code string) (string, error) {
	email = emails.Normalize(email)
	valid, err := a.registerOtpCache.Verify(ctx, email, code)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", apperrors.ErrInvalidOrExpiredOtp
	}

	registrationToken, err := tokens.GetRefreshToken()
	if err != nil {
		return "", err
	}
	if err := a.registrationTokenCache.Put(ctx, registrationToken, email); err != nil {
		return "", err
	}
	return registrationToken, nil
}

func (a *AuthService) CompleteRegistration(ctx context.Context, req *auth.CompleteRegistrationRequest) (*AuthResult, error) {
	email, err := a.registrationTokenCache.Get(ctx, req.GetRegistrationToken())
	if err != nil {
		return nil, apperrors.ErrRegistrationTokenExpired
	}

	if !req.GetAgreeToTerms() {
		return nil, apperrors.ErrTermsNotAccepted
	}

	now := time.Now().UTC()
	dateOfBirth := req.GetDateOfBirth().AsTime()
	if dateOfBirth.AddDate(minimumAge, 0, 0).After(now) {
		return nil, apperrors.ErrUnderAge
	}

	credential, err := a.credentialRepository.Create(ctx, &models.UserCredential{
		Id:    snowflake.Generate(),
		Email: email,
		Role:  models.RoleUser,
	})
	if err != nil {
		return nil, err
	}

	if err := a.publish(ctx, topics.USER_REGISTERED_TOPIC, events.UserRegisteredEvent{
		UserId:              credential.Id,
		Email:               email,
		FirstName:           req.GetFirstName(),
		Surname:             req.GetSurname(),
		DateOfBirth:         dateOfBirth,
		PhoneNumber:         req.GetPhoneNumber(),
		Gender:              req.GetGender(),
		SubscribeNewsletter: req.GetSubscribeNewsletter(),
		TermsAcceptedAt:     now,
	}); err != nil {
		return nil, err
	}

	_ = a.registrationTokenCache.Delete(ctx, req.GetRegistrationToken())

	pair, err := a.tokenService.IssueTokens(ctx, credential)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Credential: credential, TokenPair: pair}, nil
}

func (a *AuthService) RequestLoginOtp(ctx context.Context, email string) error {
	email = emails.Normalize(email)
	if _, err := a.credentialRepository.GetByEmail(ctx, email); err != nil {
		return apperrors.ErrAccountNotFound
	}
	return a.sendOtp(ctx, a.loginOtpCache, email)
}

func (a *AuthService) VerifyLoginOtp(ctx context.Context, email, code string) (*AuthResult, error) {
	email = emails.Normalize(email)
	valid, err := a.loginOtpCache.Verify(ctx, email, code)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, apperrors.ErrInvalidOrExpiredOtp
	}

	credential, err := a.credentialRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.ErrAccountNotFound
	}

	pair, err := a.tokenService.IssueTokens(ctx, credential)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Credential: credential, TokenPair: pair}, nil
}

// FindUserByEmail looks up a registered account by email. found is false
// (with a nil error) when no account uses the address.
func (a *AuthService) FindUserByEmail(ctx context.Context, email string) (*models.UserCredential, bool, error) {
	credential, err := a.credentialRepository.GetByEmail(ctx, emails.Normalize(email))
	if errors.Is(err, apperrors.ErrInvalidCredentials) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return credential, true, nil
}

func (a *AuthService) SetOrganizerStatus(ctx context.Context, userId string, isOrganizer bool) error {
	if isOrganizer {
		return a.credentialRepository.UpdateRole(ctx, userId, models.RoleUser, models.RoleOrganizer)
	}
	return a.credentialRepository.UpdateRole(ctx, userId, models.RoleOrganizer, models.RoleUser)
}

func (a *AuthService) sendOtp(ctx context.Context, cache *otp.Cache, email string) error {
	onCooldown, err := cache.OnCooldown(ctx, email)
	if err != nil {
		return err
	}
	if onCooldown {
		return apperrors.ErrOtpResendCooldown
	}

	code := otp.Generate()
	if err := cache.Put(ctx, email, code); err != nil {
		return err
	}

	return a.publish(ctx, topics.OTP_REQUESTED_TOPIC, events.OtpRequestedEvent{
		Email:   email,
		Otp:     code,
		Purpose: events.OtpPurposeAccount,
	})
}

func (a *AuthService) publish(ctx context.Context, topic string, event any) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		l.Logger.Error("Failed to marshal event for " + topic)
		return err
	}
	return a.publisher.Publish(ctx, topic, eventBytes)
}
