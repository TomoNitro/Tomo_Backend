package entity

import "time"

type DashboardSummary struct {
	ID               string    `gorm:"column:id;primaryKey;type:uuid"`
	ChildID          string    `gorm:"column:child_id;type:uuid"`
	Summary          string    `gorm:"column:summary"`
	PerformanceLevel string    `gorm:"column:performance_level"`
	Suggestion       string    `gorm:"column:suggestion"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (d *DashboardSummary) TableName() string {
	return "dashboard_summary"
}
