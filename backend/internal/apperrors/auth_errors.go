package apperrors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUnderAge                 = status.Error(codes.InvalidArgument, "You must be at least 18 years old.")
	ErrTermsNotAccepted         = status.Error(codes.InvalidArgument, "You must agree to the Terms of Service.")
	ErrEmailAlreadyExists       = status.Error(codes.AlreadyExists, "This email is already registered. Please log in instead.")
	ErrAccountNotFound          = status.Error(codes.NotFound, "No account is registered with this email.")
	ErrInvalidCredentials       = status.Error(codes.Unauthenticated, "Invalid credentials.")
	ErrInvalidRefreshToken      = status.Error(codes.Unauthenticated, "Invalid refresh token.")
	ErrRefreshTokenExpired      = status.Error(codes.Unauthenticated, "Refresh token expired.")
	ErrRefreshTokenReuse        = status.Error(codes.Unauthenticated, "Refresh token reuse detected. All sessions revoked.")
	ErrInvalidOrExpiredOtp      = status.Error(codes.InvalidArgument, "Invalid or expired OTP.")
	ErrOtpResendCooldown        = status.Error(codes.ResourceExhausted, "Please wait before requesting another OTP.")
	ErrRegistrationTokenExpired = status.Error(codes.FailedPrecondition, "Your registration session has expired. Please register again.")
)
