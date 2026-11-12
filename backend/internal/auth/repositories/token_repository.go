package repositories

import (
	"context"

	"github.com/HappyNaCl/new-debyut/internal/auth/models"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, error) {
	err := gorm.G[models.RefreshToken](r.db).Create(ctx, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *TokenRepository) GetByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var result models.RefreshToken
	err := r.db.WithContext(ctx).Preload("UserCredential").Where("token = ?", token).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *TokenRepository) Delete(ctx context.Context, token string) error {
	_, err := gorm.G[models.RefreshToken](r.db).Where("token = ?", token).Delete(ctx)
	return err
}

func (r *TokenRepository) MarkAsUsed(ctx context.Context, token string) error {
	_, err := gorm.G[models.RefreshToken](r.db).Where("token = ?", token).Update(ctx, "is_used", true)
	return err
}

func (r *TokenRepository) DeleteAllByUserId(ctx context.Context, userId string) error {
	_, err := gorm.G[models.RefreshToken](r.db).Where("user_id = ?", userId).Delete(ctx)
	return err
}
