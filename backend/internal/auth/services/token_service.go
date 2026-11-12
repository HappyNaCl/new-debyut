package services

import (
	"context"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/auth/models"
	"github.com/HappyNaCl/new-debyut/internal/auth/repositories"
	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	"github.com/HappyNaCl/new-debyut/internal/common/tokens"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenService struct {
	tokenRepository *repositories.TokenRepository
}

func NewTokenService(repo *repositories.TokenRepository) *TokenService {
	return &TokenService{
		tokenRepository: repo,
	}
}

func (s *TokenService) IssueTokens(ctx context.Context, cred *models.UserCredential) (*TokenPair, error) {
	accessToken, err := tokens.GetAccessToken(cred.Id, cred.Email, string(cred.Role))
	if err != nil {
		return nil, err
	}

	rawRefresh, err := tokens.GetRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshToken := &models.RefreshToken{
		Id:        snowflake.Generate(),
		UserId:    cred.Id,
		Token:     rawRefresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
	}

	if _, err := s.tokenRepository.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}

func (s *TokenService) RevokeToken(ctx context.Context, token string) error {
	return s.tokenRepository.Delete(ctx, token)
}

func (s *TokenService) RevokeAllTokens(ctx context.Context, userId string) error {
	return s.tokenRepository.DeleteAllByUserId(ctx, userId)
}

func (s *TokenService) RotateToken(ctx context.Context, oldToken string) (*TokenPair, error) {
	existing, err := s.tokenRepository.GetByToken(ctx, oldToken)
	if err != nil {
		return nil, apperrors.ErrInvalidRefreshToken
	}

	if existing.IsUsed {
		_ = s.tokenRepository.DeleteAllByUserId(ctx, existing.UserId)
		return nil, apperrors.ErrRefreshTokenReuse
	}

	if time.Now().After(existing.ExpiresAt) {
		return nil, apperrors.ErrRefreshTokenExpired
	}

	if err := s.tokenRepository.MarkAsUsed(ctx, oldToken); err != nil {
		return nil, err
	}

	cred := &models.UserCredential{
		Id:    existing.UserId,
		Role:  existing.UserCredential.Role,
		Email: existing.UserCredential.Email,
	}

	return s.IssueTokens(ctx, cred)
}
