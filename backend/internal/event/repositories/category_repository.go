package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/event/models"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func (r *CategoryRepository) List(ctx context.Context) ([]models.Category, error) {
	return gorm.G[models.Category](r.db).Order("name ASC").Find(ctx)
}

func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	return mapDBError(gorm.G[models.Category](r.db).Create(ctx, category))
}

func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	result := r.db.WithContext(ctx).Model(&models.Category{}).
		Where("id = ?", category.Id).
		Updates(map[string]any{"name": category.Name, "slug": category.Slug, "icon": category.Icon})
	if result.Error != nil {
		return mapDBError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrCategoryNotFound
	}
	return r.db.WithContext(ctx).Where("id = ?", category.Id).First(category).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	rows, err := gorm.G[models.Category](r.db).Where("id = ?", id).Delete(ctx)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return apperrors.ErrCategoryInUse
	}
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperrors.ErrCategoryNotFound
	}
	return nil
}

func mapDBError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "slug") {
			return apperrors.ErrCategorySlugTaken
		}
		if strings.Contains(pgErr.ConstraintName, "name") {
			return apperrors.ErrCategoryNameTaken
		}
	}
	return err
}

func (r *CategoryRepository) Exists(ctx context.Context, id string) (bool, error) {
	count, err := gorm.G[models.Category](r.db).Where("id = ?", id).Count(ctx, "id")
	return count > 0, err
}
