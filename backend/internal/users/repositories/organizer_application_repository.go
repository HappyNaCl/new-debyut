package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrganizerApplicationRepository struct {
	db *gorm.DB
}

func NewOrganizerApplicationRepository(db *gorm.DB) *OrganizerApplicationRepository {
	return &OrganizerApplicationRepository{
		db: db,
	}
}

func (r *OrganizerApplicationRepository) Create(ctx context.Context, application *models.OrganizerApplication) error {
	err := organizationConflictError(gorm.G[models.OrganizerApplication](r.db).Create(ctx, application))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperrors.ErrApplicationAlreadyPending
	}
	return err
}

func (r *OrganizerApplicationRepository) HasPending(ctx context.Context, userId string) (bool, error) {
	count, err := gorm.G[models.OrganizerApplication](r.db).
		Where("user_id = ? AND status = ?", userId, models.ApplicationStatusPending).
		Count(ctx, "id")
	return count > 0, err
}

func (r *OrganizerApplicationRepository) GetById(ctx context.Context, id string) (*models.OrganizerApplication, error) {
	application, err := gorm.G[models.OrganizerApplication](r.db).Where("id = ?", id).First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrApplicationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *OrganizerApplicationRepository) ListByUser(ctx context.Context, userId string) ([]models.OrganizerApplication, error) {
	return gorm.G[models.OrganizerApplication](r.db).
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Find(ctx)
}

func (r *OrganizerApplicationRepository) List(ctx context.Context, status models.ApplicationStatus, page, pageSize int) ([]models.OrganizerApplication, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.OrganizerApplication{}).Session(&gorm.Session{})
	if status != "" {
		query = query.Where("status = ?", status).Session(&gorm.Session{})
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var applications []models.OrganizerApplication
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&applications).Error
	return applications, total, err
}

func (r *OrganizerApplicationRepository) Decline(ctx context.Context, id, reviewerId string) (*models.OrganizerApplication, error) {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&models.OrganizerApplication{}).
		Where("id = ? AND status = ?", id, models.ApplicationStatusPending).
		Updates(map[string]any{"status": models.ApplicationStatusDeclined, "reviewed_by": reviewerId, "reviewed_at": now})
	if result.Error != nil {
		return nil, result.Error
	}

	application, err := r.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.ErrApplicationAlreadyReviewed
	}
	return application, nil
}

func (r *OrganizerApplicationRepository) Approve(ctx context.Context, id, reviewerId string, organization *models.Organization) (*models.OrganizerApplication, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var application models.OrganizerApplication
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&application).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrApplicationNotFound
			}
			return err
		}
		if application.Status != models.ApplicationStatusPending {
			return apperrors.ErrApplicationAlreadyReviewed
		}

		organization.Name = application.OrganizationName
		organization.Email = application.OrganizationEmail
		organization.Description = application.Description
		organization.AvatarUrl = application.AvatarUrl
		if err := tx.Create(organization).Error; err != nil {
			return organizationConflictError(err)
		}

		owner := &models.OrganizationMember{
			OrganizationId: organization.Id,
			UserId:         application.UserId,
			Role:           models.MemberRoleOwner,
			Status:         models.MemberStatusActive,
		}
		if err := tx.Omit("Organization").Create(owner).Error; err != nil {
			return err
		}

		return tx.Model(&models.OrganizerApplication{}).Where("id = ?", id).Updates(map[string]any{
			"status":          models.ApplicationStatusApproved,
			"reviewed_by":     reviewerId,
			"reviewed_at":     time.Now().UTC(),
			"organization_id": organization.Id,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}
