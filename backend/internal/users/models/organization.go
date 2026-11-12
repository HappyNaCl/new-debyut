package models

import "time"

type MemberRole string

const (
	MemberRoleOwner MemberRole = "OWNER"
	MemberRoleStaff MemberRole = "STAFF"
)

type Organization struct {
	Id          string `gorm:"type:varchar(20);primaryKey"`
	Name        string `gorm:"type:varchar(150);not null"`
	Email       string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text;not null"`
	AvatarUrl   string `gorm:"type:varchar(512)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrganizationMember struct {
	OrganizationId string       `gorm:"type:varchar(20);primaryKey"`
	UserId         string       `gorm:"type:varchar(20);primaryKey;index"`
	Role           MemberRole   `gorm:"type:varchar(10);not null"`
	Organization   Organization `gorm:"foreignKey:OrganizationId;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time
}

type MemberProfile struct {
	OrganizationId string
	UserId         string
	FirstName      string
	Surname        string
	Role           MemberRole
	CreatedAt      time.Time
}

type MembershipOrganization struct {
	Organization
	Role MemberRole
}

type OrganizationSummary struct {
	Organization
	MemberCount int64
}
