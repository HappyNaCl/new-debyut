package events

const (
	OtpPurposeAccount           = "account"
	OtpPurposeOrganizationEmail = "organization-email"
)

type OtpRequestedEvent struct {
	Email   string
	Otp     string
	Purpose string
}
