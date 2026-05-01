package entity

type FinanceTheme struct {
	ID    string `gorm:"column:id;primaryKey"`
	Topic string `gorm:"column:topic"`
}

func (f *FinanceTheme) TableName() string {
	return "finance_theme"
}
