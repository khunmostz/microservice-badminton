package domain

type Court struct {
	ID           string `gorm:"primaryKey"`
	Venue        string
	CourtNo      int32
	PricePerHour int64
	OpenFrom     string // HH:mm
	OpenTo       string // HH:mm
	OwnerID      string // จาก JWT (role OWNER/ADMIN)
}
