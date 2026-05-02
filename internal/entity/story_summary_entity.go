package entity

import "time"

type StorySummary struct {
	ID          string    `gorm:"column:id;primaryKey;type:uuid"`
	Title       string    `gorm:"column:title"`
	Description string    `gorm:"column:description"`
	Performance string    `gorm:"column:performance"`
	CreatedAt   time.Time `gorm:"-"`
}

func (s *StorySummary) TableName() string {
	return "story_summary"
}
