package domain

import "time"

type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex"`
	Name      string
	Phone     string
	AvatarURL string
	Role      string `gorm:"index"` // USER|OWNER|ADMIN
	CreatedAt time.Time
	UpdatedAt time.Time
}
