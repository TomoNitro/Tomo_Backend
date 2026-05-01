package entity

import "time"

type Children struct {
	ID        string     `gorm:"column:id;primaryKey"`
	ParentId  string     `gorm:"column:parent_id"`
	Name      string     `gorm:"column:name"`
	Pin       string     `gorm:"column:pin"`
	LoginAt   *time.Time `gorm:"column:login_at"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (c *Children) TableName() string {
	return "children"
}
