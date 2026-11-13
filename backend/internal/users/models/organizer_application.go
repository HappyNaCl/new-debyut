package models

import "time"

type ApplicationStatus string

const (
	ApplicationStatusPending  ApplicationStatus = "PENDING"
	ApplicationStatusApproved ApplicationStatus = "APPROVED"
	ApplicationStatusDeclined ApplicationStatus = "DECLINED"
)

type OrganizerApplication struct {
	Id                string            `gorm:"type:varchar(20);primaryKey"`
	UserId            string            `gorm:"type:varchar(20);not null;index;uniqueIndex:idx_organizer_applications_one_pending,where:status = 'PENDING'"`
	OrganizationName  string            `gorm:"type:varchar(150);not null;uniqueIndex:idx_organizer_applications_pending_name,expression:lower(organization_name),where:status = 'PENDING'"`
	OrganizationEmail string            `gorm:"type:varchar(255);not null;uniqueIndex:idx_organizer_applications_pending_email,where:status = 'PENDING'"`
	Description       string            `gorm:"type:text;not null"`
	AvatarUrl         string            `gorm:"type:varchar(512)"`
	Status            ApplicationStatus `gorm:"type:varchar(10);not null;default:PENDING;index"`
	ReviewedBy        *string           `gorm:"type:varchar(20)"`
	ReviewedAt        *time.Time
	OrganizationId    *string `gorm:"type:varchar(20)"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
