package services

import (
	"context"

	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
)

// Account is the part of a registered debyut account the users service needs.
// Emails and roles live in the auth service, so this is looked up over gRPC.
type Account struct {
	Id      string
	IsAdmin bool
}

// AccountDirectory finds registered accounts by email.
type AccountDirectory interface {
	// FindByEmail returns nil (and a nil error) when no account uses the email.
	FindByEmail(ctx context.Context, email string) (*Account, error)
}

type authAccountDirectory struct {
	client auth.AuthServiceClient
}

func NewAuthAccountDirectory(client auth.AuthServiceClient) AccountDirectory {
	return &authAccountDirectory{client: client}
}

func (d *authAccountDirectory) FindByEmail(ctx context.Context, email string) (*Account, error) {
	res, err := d.client.FindUserByEmail(ctx, &auth.EmailRequest{Email: email})
	if err != nil {
		return nil, err
	}
	if !res.GetFound() {
		return nil, nil
	}
	return &Account{Id: res.GetId(), IsAdmin: res.GetRole() == auth.Role_ADMIN}, nil
}
