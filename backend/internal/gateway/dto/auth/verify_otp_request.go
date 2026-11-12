package dto

type VerifyOtpRequest struct {
	Email string `json:"email" binding:"required,email"`
	Otp   string `json:"otp"   binding:"required,len=6,numeric"`
}
