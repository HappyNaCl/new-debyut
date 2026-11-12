package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{
		db: db,
	}
}

func (r *InvitationRepository) Create(ctx context.Context, invitation *models.OrganizationInvitation) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.OrganizationInvitation{}).
			Where("organization_id = ? AND email = ? AND status = ? AND expires_at <= ?", invitation.OrganizationId, invitation.Email, models.InvitationStatusPending, time.Now().UTC()).
			Update("status", models.InvitationStatusExpired).Error; err != nil {
			return err
		}

		err := tx.Omit("Organization").Create(invitation).Error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrInvitationAlreadyPending
		}
		return err
	})
}

func (r *InvitationRepository) ListPendingByOrganization(ctx context.Context, organizationId string) ([]models.OrganizationInvitation, error) {
	var invitations []models.OrganizationInvitation
	err := r.db.WithContext(ctx).
		Preload("Organization").
		Where("organization_id = ? AND status = ? AND expires_at > ?", organizationId, models.InvitationStatusPending, time.Now().UTC()).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

func (r *InvitationRepository) ListPendingByEmail(ctx context.Context, email string) ([]models.OrganizationInvitation, error) {
	var invitations []models.OrganizationInvitation
	err := r.db.WithContext(ctx).
		Preload("Organization").
		Where("email = ? AND status = ? AND expires_at > ?", strings.ToLower(email), models.InvitationStatusPending, time.Now().UTC()).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

func (r *InvitationRepository) Revoke(ctx context.Context, id, organizationId string) error {
	result := r.db.WithContext(ctx).Model(&models.OrganizationInvitation{}).
		Where("id = ? AND organization_id = ? AND status = ?", id, organizationId, models.InvitationStatusPending).
		Update("status", models.InvitationStatusRevoked)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrInvitationNotFound
	}
	return nil
}

func (r *InvitationRepository) Decline(ctx context.Context, id, email string) error {
	result := r.db.WithContext(ctx).Model(&models.OrganizationInvitation{}).
		Where("id = ? AND email = ? AND status = ? AND expires_at > ?", id, strings.ToLower(email), models.InvitationStatusPending, time.Now().UTC()).
		Update("status", models.InvitationStatusDeclined)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrInvitationNotFound
	}
	return nil
}

func (r *InvitationRepository) Accept(ctx context.Context, id, userId, email string) (*models.OrganizationInvitation, error) {
	var invitation models.OrganizationInvitation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND email = ?", id, strings.ToLower(email)).
			First(&invitation).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrInvitationNotFound
		}
		if err != nil {
			return err
		}
		if invitation.Status != models.InvitationStatusPending || !invitation.ExpiresAt.After(time.Now().UTC()) {
			return apperrors.ErrInvitationNotPending
		}

		err = tx.Omit("Organization").Create(&models.OrganizationMember{
			OrganizationId: invitation.OrganizationId,
			UserId:         userId,
			Role:           invitation.Role,
		}).Error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrAlreadyMember
		}
		if err != nil {
			return err
		}

		invitation.Status = models.InvitationStatusAccepted
		return tx.Model(&invitation).Update("status", models.InvitationStatusAccepted).Error
	})
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}
