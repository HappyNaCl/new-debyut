package dto

type OrganizationEmailRequest struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

type VerifyOrganizationEmailRequest struct {
	Email string `json:"email" binding:"required,email,max=255"`
	Otp   string `json:"otp"   binding:"required,len=6,numeric"`
}

type VerificationTokenResponse struct {
	VerificationToken string `json:"verificationToken"`
}

type SubmitApplicationRequest struct {
	VerificationToken string `json:"verificationToken" binding:"required"`
	OrganizationName  string `json:"organizationName"  binding:"required,max=150"`
	Description       string `json:"description"       binding:"required,max=2000"`
}
