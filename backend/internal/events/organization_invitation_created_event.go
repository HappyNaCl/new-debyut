package events

import "time"

type OrganizationInvitationCreatedEvent struct {
	InvitationId     string
	OrganizationName string
	Email            string
	Role             string
	InviterName      string
	ExpiresAt        time.Time
}
