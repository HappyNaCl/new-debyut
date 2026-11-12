package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/users/models"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	err := gorm.G[models.User](r.db).Create(ctx, user)
	if err != nil {
		return nil, mapDBError(err)
	}
	return user, nil
}

func (r *UserRepository) GetById(ctx context.Context, id string) (*models.User, error) {
	user, err := gorm.G[models.User](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func mapDBError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "email") {
			return apperrors.ErrEmailAlreadyExists
		}
	}
	return err
}
