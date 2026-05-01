package entity

import "time"

type ChildProgress struct {
	ChildID   string    `gorm:"column:child_id;primaryKey;type:uuid"`
	TotalExp  int       `gorm:"column:total_exp"`
	Level     int       `gorm:"column:level"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (c *ChildProgress) TableName() string {
	return "child_progress"
}
