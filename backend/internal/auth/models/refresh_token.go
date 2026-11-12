package models

import "time"

type RefreshToken struct {
	Id        string    `gorm:"type:varchar(20);primaryKey"`
	UserId    string    `gorm:"type:varchar(20);not null"`
	Token     string    `gorm:"type:text;not null"`
	IsUsed    bool      `gorm:"not null;default:false"`
	ExpiresAt time.Time `gorm:"not null"`

	UserCredential UserCredential `gorm:"foreignKey:UserId"`
}
