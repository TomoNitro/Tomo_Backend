package entity

import "time"

type Decision struct {
	DecisionID string    `gorm:"column:decision_id;primaryKey"`
	SessionID  string    `gorm:"column:session_id"`
	ChildID    string    `gorm:"column:child_id"`
	NodeID     string    `gorm:"column:node_id"`
	IsWise     bool      `gorm:"column:is_wise"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (d *Decision) TableName() string {
	return "decisions"
}
