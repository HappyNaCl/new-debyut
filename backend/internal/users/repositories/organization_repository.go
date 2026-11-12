package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{
		db: db,
	}
}

func (r *OrganizationRepository) ListByMember(ctx context.Context, userId string) ([]models.MembershipOrganization, error) {
	var organizations []models.MembershipOrganization
	err := r.db.WithContext(ctx).
		Table("organizations").
		Select("organizations.*, organization_members.role").
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ?", userId).
		Order("organizations.created_at DESC").
		Scan(&organizations).Error
	return organizations, err
}

func (r *OrganizationRepository) GetForMember(ctx context.Context, organizationId, userId string) (*models.MembershipOrganization, error) {
	var organizations []models.MembershipOrganization
	err := r.db.WithContext(ctx).
		Table("organizations").
		Select("organizations.*, organization_members.role").
		Joins("JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organizations.id = ? AND organization_members.user_id = ?", organizationId, userId).
		Limit(1).
		Scan(&organizations).Error
	if err != nil {
		return nil, err
	}
	if len(organizations) == 0 {
		return nil, apperrors.ErrOrganizationNotFound
	}
	return &organizations[0], nil
}

func (r *OrganizationRepository) Update(ctx context.Context, organizationId, name, description string) error {
	return r.db.WithContext(ctx).
		Model(&models.Organization{}).
		Where("id = ?", organizationId).
		Updates(map[string]any{"name": name, "description": description}).Error
}

func (r *OrganizationRepository) UpdateAvatar(ctx context.Context, organizationId, avatarUrl string) error {
	return r.db.WithContext(ctx).
		Model(&models.Organization{}).
		Where("id = ?", organizationId).
		Update("avatar_url", avatarUrl).Error
}

func (r *OrganizationRepository) ListMembers(ctx context.Context, organizationId string) ([]models.MemberProfile, error) {
	var members []models.MemberProfile
	err := r.db.WithContext(ctx).
		Table("organization_members").
		Select("organization_members.organization_id, organization_members.user_id, users.first_name, users.surname, organization_members.role, organization_members.created_at").
		Joins("LEFT JOIN users ON users.id = organization_members.user_id").
		Where("organization_members.organization_id = ?", organizationId).
		Order("organization_members.created_at ASC").
		Scan(&members).Error
	return members, err
}

func (r *OrganizationRepository) GetMember(ctx context.Context, organizationId, userId string) (*models.OrganizationMember, error) {
	member, err := gorm.G[models.OrganizationMember](r.db).
		Where("organization_id = ? AND user_id = ?", organizationId, userId).
		First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *OrganizationRepository) List(ctx context.Context, search string, page, pageSize int) ([]models.OrganizationSummary, int64, error) {
	query := r.db.WithContext(ctx).Table("organizations").Session(&gorm.Session{})
	if search != "" {
		pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(search) + "%"
		query = query.Where("organizations.name ILIKE ? OR organizations.email ILIKE ?", pattern, pattern).Session(&gorm.Session{})
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var organizations []models.OrganizationSummary
	err := query.
		Select("organizations.*, (SELECT COUNT(*) FROM organization_members WHERE organization_members.organization_id = organizations.id) AS member_count").
		Order("organizations.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&organizations).Error
	return organizations, total, err
}

func (r *OrganizationRepository) RemoveMember(ctx context.Context, organizationId, userId string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var members []models.OrganizationMember
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("organization_id = ?", organizationId).
			Find(&members).Error; err != nil {
			return err
		}

		var target *models.OrganizationMember
		owners := 0
		for index := range members {
			if members[index].Role == models.MemberRoleOwner {
				owners++
			}
			if members[index].UserId == userId {
				target = &members[index]
			}
		}
		if target == nil {
			return apperrors.ErrMemberNotFound
		}
		if target.Role == models.MemberRoleOwner && owners <= 1 {
			return apperrors.ErrLastOwner
		}

		return tx.Where("organization_id = ? AND user_id = ?", organizationId, userId).
			Delete(&models.OrganizationMember{}).Error
	})
}

func (r *OrganizationRepository) CountMemberships(ctx context.Context, userId string) (int64, error) {
	return gorm.G[models.OrganizationMember](r.db).Where("user_id = ?", userId).Count(ctx, "user_id")
}
