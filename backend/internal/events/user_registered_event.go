package events

import "time"

type UserRegisteredEvent struct {
	UserId              string
	Email               string
	FirstName           string
	Surname             string
	DateOfBirth         time.Time
	PhoneNumber         string
	Gender              string
	SubscribeNewsletter bool
	TermsAcceptedAt     time.Time
}
