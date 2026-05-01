package entity

import "time"

type CoinTransaction struct {
	ID        string    `gorm:"column:id;primaryKey;type:uuid"`
	ChildID   string    `gorm:"column:child_id;type:uuid"`
	Amount    int       `gorm:"column:amount"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	Child     *Children `gorm:"foreignKey:ChildID;references:ID"`
}

func (c *CoinTransaction) TableName() string {
	return "coin_transactions"
}
