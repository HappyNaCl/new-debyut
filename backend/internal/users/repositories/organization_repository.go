package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/jackc/pgx/v5/pgconn"
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
		Where("organization_members.user_id = ? AND organization_members.status = ?", userId, models.MemberStatusActive).
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
		Where("organizations.id = ? AND organization_members.user_id = ? AND organization_members.status = ?", organizationId, userId, models.MemberStatusActive).
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
	err := r.db.WithContext(ctx).
		Model(&models.Organization{}).
		Where("id = ?", organizationId).
		Updates(map[string]any{"name": name, "description": description}).Error
	return organizationConflictError(err)
}

// NameInUse reports whether another organization, or a pending organizer
// application, already uses the name (case-insensitive).
func (r *OrganizationRepository) NameInUse(ctx context.Context, name, excludeOrganizationId string) (bool, error) {
	var inUse bool
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXISTS (SELECT 1 FROM organizations WHERE lower(name) = lower(?) AND id <> ?)
		    OR EXISTS (SELECT 1 FROM organizer_applications WHERE lower(organization_name) = lower(?) AND status = ?)`,
		name, excludeOrganizationId, name, models.ApplicationStatusPending,
	).Scan(&inUse).Error
	return inUse, err
}

// EmailInUse reports whether an organization, or a pending organizer
// application, already uses the email. Emails are stored normalized (see
// emails.Normalize), so callers must pass a normalized email.
func (r *OrganizationRepository) EmailInUse(ctx context.Context, email string) (bool, error) {
	var inUse bool
	err := r.db.WithContext(ctx).Raw(`
		SELECT EXISTS (SELECT 1 FROM organizations WHERE email = ?)
		    OR EXISTS (SELECT 1 FROM organizer_applications WHERE organization_email = ? AND status = ?)`,
		email, email, models.ApplicationStatusPending,
	).Scan(&inUse).Error
	return inUse, err
}

// organizationConflictError maps unique-index violations on organization
// names and emails (organizations and pending applications) to app errors.
func organizationConflictError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return err
	}
	switch pgErr.ConstraintName {
	case "idx_organizations_name_lower", "idx_organizer_applications_pending_name":
		return apperrors.ErrOrganizationNameTaken
	case "idx_organizations_email", "idx_organizer_applications_pending_email":
		return apperrors.ErrOrganizationEmailTaken
	}
	return err
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
		Select("organization_members.organization_id, organization_members.user_id, users.first_name, users.surname, organization_members.role, organization_members.status, organization_members.created_at").
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
			if isActiveOwner(&members[index]) {
				owners++
			}
			if members[index].UserId == userId {
				target = &members[index]
			}
		}
		if target == nil {
			return apperrors.ErrMemberNotFound
		}
		if isActiveOwner(target) && owners <= 1 {
			return apperrors.ErrLastOwner
		}

		return tx.Where("organization_id = ? AND user_id = ?", organizationId, userId).
			Delete(&models.OrganizationMember{}).Error
	})
}

// SetMemberStatus activates or deactivates a member. The last active owner
// cannot be deactivated.
func (r *OrganizationRepository) SetMemberStatus(ctx context.Context, organizationId, userId string, status models.MemberStatus) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var members []models.OrganizationMember
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("organization_id = ?", organizationId).
			Find(&members).Error; err != nil {
			return err
		}

		var target *models.OrganizationMember
		activeOwners := 0
		for index := range members {
			if isActiveOwner(&members[index]) {
				activeOwners++
			}
			if members[index].UserId == userId {
				target = &members[index]
			}
		}
		if target == nil {
			return apperrors.ErrMemberNotFound
		}
		if status == models.MemberStatusInactive && isActiveOwner(target) && activeOwners <= 1 {
			return apperrors.ErrLastActiveOwner
		}

		return tx.Model(&models.OrganizationMember{}).
			Where("organization_id = ? AND user_id = ?", organizationId, userId).
			Update("status", status).Error
	})
}

// CountMemberships counts the organizations where the user is an active member.
func (r *OrganizationRepository) CountMemberships(ctx context.Context, userId string) (int64, error) {
	return gorm.G[models.OrganizationMember](r.db).
		Where("user_id = ? AND status = ?", userId, models.MemberStatusActive).
		Count(ctx, "user_id")
}

func isActiveOwner(member *models.OrganizationMember) bool {
	return member.Role == models.MemberRoleOwner && member.Status == models.MemberStatusActive
}
