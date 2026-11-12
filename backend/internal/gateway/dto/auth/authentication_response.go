package dto

import (
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
)

type AuthResponse struct {
	Id           string `json:"id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"-"`
}

func NewAuthResponse(res *auth.AuthenticationResponse) *AuthResponse {
	return &AuthResponse{
		Id:           res.GetId(),
		Email:        res.GetEmail(),
		Role:         res.Role.String(),
		AccessToken:  res.GetAccessToken(),
		RefreshToken: res.GetRefreshToken(),
	}
}
