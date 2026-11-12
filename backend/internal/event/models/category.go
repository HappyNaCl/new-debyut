package models

import "time"

type Category struct {
	Id        string `gorm:"type:varchar(20);primaryKey"`
	Slug      string `gorm:"type:varchar(64);not null;uniqueIndex:idx_categories_slug"`
	Name      string `gorm:"type:varchar(100);not null;uniqueIndex:idx_categories_name"`
	Icon      string `gorm:"type:varchar(64);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
