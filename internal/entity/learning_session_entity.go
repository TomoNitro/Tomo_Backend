package entity

import "time"

type LearningSession struct {
	SessionID     string     `gorm:"column:session_id;primaryKey"`
	ChildID       string     `gorm:"column:child_id"`
	StoryID       string     `gorm:"column:story_id"`
	StartedAt     time.Time  `gorm:"column:started_at"`
	CompletedAt   *time.Time `gorm:"column:completed_at"`
}

func (l *LearningSession) TableName() string {
	return "learning_sessions"
}
