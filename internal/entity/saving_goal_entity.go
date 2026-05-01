package entity

type SavingGoal struct {
	ID          string `gorm:"column:id;primaryKey;type:uuid"`
	ChildID     string `gorm:"column:child_id;type:uuid;uniqueIndex"`
	MarketID    string `gorm:"column:market_id;type:uuid"`
	GoalName    string `gorm:"column:goal_name"`
	TargetCoin  int    `gorm:"column:target_coin"`
	CurrentCoin int    `gorm:"column:current_coin"`

	Child  *Children `gorm:"foreignKey:ChildID;references:ID"`
	Market *Market   `gorm:"foreignKey:MarketID;references:ID"`
}

func (s *SavingGoal) TableName() string {
	return "saving_goals"
}
