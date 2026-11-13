package handlers

import (
	"context"

	"github.com/HappyNaCl/new-debyut/internal/auth/services"
	"github.com/HappyNaCl/new-debyut/internal/genproto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthGrpcHandler struct {
	authSvc  *services.AuthService
	tokenSvc *services.TokenService

	auth.UnimplementedAuthServiceServer
}

func NewAuthGrpcHandler(grpc *grpc.Server, authSvc *services.AuthService, tokenSvc *services.TokenService) {
	handler := &AuthGrpcHandler{
		authSvc:  authSvc,
		tokenSvc: tokenSvc,
	}

	auth.RegisterAuthServiceServer(grpc, handler)
}

func (h *AuthGrpcHandler) RequestRegisterOtp(ctx context.Context, req *auth.EmailRequest) (*emptypb.Empty, error) {
	if err := h.authSvc.RequestRegisterOtp(ctx, req.GetEmail()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *AuthGrpcHandler) VerifyRegisterOtp(ctx context.Context, req *auth.VerifyOtpRequest) (*auth.VerifyRegisterOtpResponse, error) {
	registrationToken, err := h.authSvc.VerifyRegisterOtp(ctx, req.GetEmail(), req.GetOtp())
	if err != nil {
		return nil, err
	}
	return &auth.VerifyRegisterOtpResponse{RegistrationToken: registrationToken}, nil
}

func (h *AuthGrpcHandler) CompleteRegistration(ctx context.Context, req *auth.CompleteRegistrationRequest) (*auth.AuthenticationResponse, error) {
	result, err := h.authSvc.CompleteRegistration(ctx, req)
	if err != nil {
		return nil, err
	}
	return toAuthResponse(result), nil
}

func (h *AuthGrpcHandler) RequestLoginOtp(ctx context.Context, req *auth.EmailRequest) (*emptypb.Empty, error) {
	if err := h.authSvc.RequestLoginOtp(ctx, req.GetEmail()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *AuthGrpcHandler) VerifyLoginOtp(ctx context.Context, req *auth.VerifyOtpRequest) (*auth.AuthenticationResponse, error) {
	result, err := h.authSvc.VerifyLoginOtp(ctx, req.GetEmail(), req.GetOtp())
	if err != nil {
		return nil, err
	}
	return toAuthResponse(result), nil
}

func (h *AuthGrpcHandler) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenResponse, error) {
	result, err := h.tokenSvc.RotateToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}
	return &auth.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (h *AuthGrpcHandler) Logout(ctx context.Context, req *auth.LogoutRequest) (*emptypb.Empty, error) {
	userId := req.GetUserId()
	if userId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	if err := h.tokenSvc.RevokeAllTokens(ctx, userId); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *AuthGrpcHandler) FindUserByEmail(ctx context.Context, req *auth.EmailRequest) (*auth.FindUserResponse, error) {
	credential, found, err := h.authSvc.FindUserByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	if !found {
		return &auth.FindUserResponse{}, nil
	}
	return &auth.FindUserResponse{
		Found: true,
		Id:    credential.Id,
		Role:  auth.Role(auth.Role_value[string(credential.Role)]),
	}, nil
}

func toAuthResponse(result *services.AuthResult) *auth.AuthenticationResponse {
	cred := result.Credential

	return &auth.AuthenticationResponse{
		Id:           cred.Id,
		Email:        cred.Email,
		Role:         auth.Role(auth.Role_value[string(cred.Role)]),
		AccessToken:  result.TokenPair.AccessToken,
		RefreshToken: result.TokenPair.RefreshToken,
	}
}
