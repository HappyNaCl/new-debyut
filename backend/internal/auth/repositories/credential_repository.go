package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/HappyNaCl/new-debyut/internal/apperrors"
	"github.com/HappyNaCl/new-debyut/internal/auth/models"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type CredentialRepository struct {
	db *gorm.DB
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{
		db: db,
	}
}

func (r *CredentialRepository) Create(ctx context.Context, cred *models.UserCredential) (*models.UserCredential, error) {
	err := gorm.G[models.UserCredential](r.db).Create(ctx, cred)
	if err != nil {
		return nil, mapDBError(err)
	}
	return cred, nil
}

func (r *CredentialRepository) GetByEmail(ctx context.Context, email string) (*models.UserCredential, error) {
	cred, err := gorm.G[models.UserCredential](r.db).Where("email = ?", email).First(ctx)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrInvalidCredentials
	} else if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *CredentialRepository) GetById(ctx context.Context, id string) (*models.UserCredential, error) {
	cred, err := gorm.G[models.UserCredential](r.db).Where("id = ?", id).First(ctx)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrInvalidCredentials
	} else if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *CredentialRepository) UpdateRole(ctx context.Context, id string, from, to models.Role) error {
	_, err := gorm.G[models.UserCredential](r.db).Where("id = ? AND role = ?", id, from).Update(ctx, "role", to)
	return err
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
