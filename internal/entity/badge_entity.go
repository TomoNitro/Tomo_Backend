package entity

import "time"

type Badge struct {
	ID            string    `gorm:"column:id;primaryKey;type:uuid"`
	Name          string    `gorm:"column:name"`
	Description   string    `gorm:"column:description"`
	LevelRequired int       `gorm:"column:level_required"`
	ImageURL      *string   `gorm:"column:image_url"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (b *Badge) TableName() string {
	return "badges"
}
