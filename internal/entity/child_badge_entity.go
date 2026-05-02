package entity

import "time"

type ChildBadge struct {
	ID       string    `gorm:"column:id;primaryKey;type:uuid"`
	ChildID  string    `gorm:"column:child_id;type:uuid"`
	BadgeID  string    `gorm:"column:badge_id;type:uuid"`
	EarnedAt time.Time `gorm:"column:earned_at;autoCreateTime"`
}

func (c *ChildBadge) TableName() string {
	return "child_badges"
}
