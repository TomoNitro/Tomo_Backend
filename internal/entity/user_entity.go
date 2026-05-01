package entity

import "time"

type User struct {
	ID        string     `gorm:"column:id;primaryKey"`
	Email     string     `gorm:"column:email;unique"`
	Password  string     `gorm:"column:password"`
	Username  string     `gorm:"column:username;unique"`
	CreatedAt time.Time  `gorm:"column:createdat;autoCreateTime:milli"`
	Children  []Children `gorm:"foreignKey:ParentId"`
}

func (u *User) TableName() string {
	return "users"
}
