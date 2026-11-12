package models

import "time"

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "PENDING"
	InvitationStatusAccepted InvitationStatus = "ACCEPTED"
	InvitationStatusDeclined InvitationStatus = "DECLINED"
	InvitationStatusRevoked  InvitationStatus = "REVOKED"
	InvitationStatusExpired  InvitationStatus = "EXPIRED"
)

type OrganizationInvitation struct {
	Id             string           `gorm:"type:varchar(20);primaryKey"`
	OrganizationId string           `gorm:"type:varchar(20);not null;uniqueIndex:idx_organization_invitations_one_pending,where:status = 'PENDING'"`
	Email          string           `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_organization_invitations_one_pending,where:status = 'PENDING'"`
	Role           MemberRole       `gorm:"type:varchar(10);not null"`
	Status         InvitationStatus `gorm:"type:varchar(10);not null;default:PENDING"`
	InvitedBy      string           `gorm:"type:varchar(20);not null"`
	ExpiresAt      time.Time        `gorm:"not null"`
	Organization   Organization     `gorm:"foreignKey:OrganizationId;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
