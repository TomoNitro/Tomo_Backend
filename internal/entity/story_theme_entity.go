package entity

type StoryTheme struct {
	ID        string `gorm:"column:id;primaryKey"`
	Title     string `gorm:"column:title"`
	FullStory string `gorm:"column:full_story"`
}

func (s *StoryTheme) TableName() string {
	return "story_theme"
}
