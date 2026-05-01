package entity

type StoryTheme struct {
	ID        string `gorm:"column:id;primaryKey"`
	Topic     string `gorm:"column:topic"`
	FullStory string `gorm:"column:full_story"`
}

func (s *StoryTheme) TableName() string {
	return "story_theme"
}
