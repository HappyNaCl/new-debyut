package services

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	l "github.com/HappyNaCl/new-debyut/internal/common/logger"
	"github.com/HappyNaCl/new-debyut/internal/common/snowflake"
	"github.com/HappyNaCl/new-debyut/internal/event/models"
	"github.com/HappyNaCl/new-debyut/internal/event/repositories"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	categoriesCacheKey = "categories"
	categoriesCacheTtl = time.Hour
)

var (
	slugPattern      = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	slugSeparator    = regexp.MustCompile(`[^a-z0-9]+`)
	maximumSlugRunes = 64
)

type CategoryService struct {
	categoryRepository *repositories.CategoryRepository
	redis              *redis.Client
}

func NewCategoryService(categoryRepository *repositories.CategoryRepository, redis *redis.Client) *CategoryService {
	return &CategoryService{
		categoryRepository: categoryRepository,
		redis:              redis,
	}
}

func (s *CategoryService) List(ctx context.Context) ([]models.Category, error) {
	if cached, err := s.redis.Get(ctx, categoriesCacheKey).Bytes(); err == nil {
		var categories []models.Category
		if err := json.Unmarshal(cached, &categories); err == nil {
			return categories, nil
		}
	}

	categories, err := s.categoryRepository.List(ctx)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(categories); err == nil {
		if err := s.redis.Set(ctx, categoriesCacheKey, data, categoriesCacheTtl).Err(); err != nil {
			l.Logger.Warn("failed to cache categories", zap.Error(err))
		}
	}
	return categories, nil
}

func (s *CategoryService) Create(ctx context.Context, name, slug, icon string) (*models.Category, error) {
	category, err := buildCategory(snowflake.Generate(), name, slug, icon)
	if err != nil {
		return nil, err
	}
	if err := s.categoryRepository.Create(ctx, category); err != nil {
		return nil, err
	}
	s.invalidate(ctx)
	return category, nil
}

func (s *CategoryService) Update(ctx context.Context, id, name, slug, icon string) (*models.Category, error) {
	category, err := buildCategory(id, name, slug, icon)
	if err != nil {
		return nil, err
	}
	if err := s.categoryRepository.Update(ctx, category); err != nil {
		return nil, err
	}
	s.invalidate(ctx)
	return category, nil
}

func (s *CategoryService) Delete(ctx context.Context, id string) error {
	if err := s.categoryRepository.Delete(ctx, id); err != nil {
		return err
	}
	s.invalidate(ctx)
	return nil
}

func (s *CategoryService) invalidate(ctx context.Context) {
	if err := s.redis.Del(ctx, categoriesCacheKey).Err(); err != nil {
		l.Logger.Warn("failed to invalidate categories cache", zap.Error(err))
	}
}

func buildCategory(id, name, slug, icon string) (*models.Category, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if slug == "" {
		slug = Slugify(name)
	}
	icon = strings.TrimSpace(icon)

	if name == "" || !slugPattern.MatchString(slug) || !slugPattern.MatchString(icon) {
		return nil, apperrors.ErrInvalidCategory
	}

	return &models.Category{
		Id:   id,
		Name: name,
		Slug: slug,
		Icon: icon,
	}, nil
}

func Slugify(name string) string {
	slug := strings.Trim(slugSeparator.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if runes := []rune(slug); len(runes) > maximumSlugRunes {
		slug = strings.TrimRight(string(runes[:maximumSlugRunes]), "-")
	}
	return slug
}
