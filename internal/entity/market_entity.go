package entity

import "time"

type Market struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Title     string    `gorm:"column:title"`
	ImageURL  string    `gorm:"column:image_url"`
	Price     int       `gorm:"column:price"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (m *Market) TableName() string {
	return "market"
}
