package entity

import "time"

type ExpTransaction struct {
	ID          string    `gorm:"column:id;primaryKey;type:uuid"`
	ChildID     string    `gorm:"column:child_id;type:uuid"`
	Amount      int       `gorm:"column:amount"`
	Source      string    `gorm:"column:source"`
	ReferenceID *string   `gorm:"column:reference_id;type:uuid"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (e *ExpTransaction) TableName() string {
	return "exp_transactions"
}
