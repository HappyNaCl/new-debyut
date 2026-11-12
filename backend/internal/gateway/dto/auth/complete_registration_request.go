package dto

type CompleteRegistrationRequest struct {
	RegistrationToken   string `json:"registrationToken"   binding:"required"`
	FirstName           string `json:"firstName"           binding:"required,max=100"`
	Surname             string `json:"surname"             binding:"omitempty,max=100"`
	DateOfBirth         string `json:"dateOfBirth"         binding:"required,datetime=2006-01-02"`
	PhoneNumber         string `json:"phoneNumber"         binding:"required,e164"`
	Gender              string `json:"gender"              binding:"required,oneof=MALE FEMALE"`
	AgreeToTerms        bool   `json:"agreeToTerms"        binding:"required"`
	SubscribeNewsletter bool   `json:"subscribeNewsletter"`
}
