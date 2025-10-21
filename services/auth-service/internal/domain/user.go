package domain

type Role string

const (
	RoleUser  Role = "USER"
	RoleOwner Role = "OWNER"
	RoleAdmin Role = "ADMIN"
)

type User struct {
	ID           string `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
	Name         string
	Role         Role
}
