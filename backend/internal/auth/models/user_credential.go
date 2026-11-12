package models

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin     Role = "ADMIN"
	RoleUser      Role = "USER"
	RoleOrganizer Role = "ORGANIZER"
)

type UserCredential struct {
	Id        string `gorm:"type:varchar(20);primaryKey"`
	Email     string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Role      Role   `gorm:"type:varchar(20);not null;default:USER"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
