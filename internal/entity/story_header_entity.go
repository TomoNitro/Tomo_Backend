package entity

import "time"

type StoryHeader struct {
	StoryID     string    `gorm:"column:story_id;primaryKey;autoIncrement"`
	RootStoryID *string   `gorm:"column:root_story_id"`
	SummaryID   *string   `gorm:"column:summary_id"`
	ParentID    *string   `gorm:"column:parent_id"`
	Title       string    `gorm:"column:title"`
	ImageURL    string    `gorm:"column:image_url"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`

	// Relations
	// RootStory *StoryNode    `gorm:"foreignKey:RootStoryID"`
	// Summary   *StorySummary `gorm:"foreignKey:SummaryID"`
	// Parent    *User         `gorm:"foreignKey:ParentID"`

	// LearningSessions []LearningSession `gorm:"foreignKey:StoryID"`
}

func (s *StoryHeader) TableName() string {
	return "story_header"
}
