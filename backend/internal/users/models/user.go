package models

import (
	"time"

	"gorm.io/gorm"
)

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
)

type User struct {
	Id                  string    `gorm:"type:varchar(20);primaryKey"`
	FirstName           string    `gorm:"type:varchar(100);not null"`
	Surname             string    `gorm:"type:varchar(100)"`
	DateOfBirth         time.Time `gorm:"type:date;not null"`
	PhoneNumber         string    `gorm:"type:varchar(20);not null"`
	Gender              Gender    `gorm:"type:varchar(10);not null"`
	SubscribeNewsletter bool      `gorm:"not null;default:false"`
	AvatarUrl           string    `gorm:"type:varchar(512)"`
	TermsAcceptedAt     time.Time `gorm:"not null"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}
