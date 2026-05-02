package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StoryPlayRepository struct {
	Log *zap.Logger
}

func NewStoryPlayRepository(log *zap.Logger) *StoryPlayRepository {
	return &StoryPlayRepository{
		Log: log,
	}
}

func (r *StoryPlayRepository) FindStoryHeaderByID(db *gorm.DB, storyHeader *entity.StoryHeader, storyID string) error {
	return db.Where("story_id = ?", storyID).First(storyHeader).Error
}

func (r *StoryPlayRepository) FindStoryNodeByID(db *gorm.DB, storyNode *entity.StoryNode, nodeID string) error {
	return db.Where("node_id = ?", nodeID).First(storyNode).Error
}

func (r *StoryPlayRepository) CreateLearningSession(db *gorm.DB, session *entity.LearningSession) error {
	return db.Create(session).Error
}

func (r *StoryPlayRepository) FindLearningSessionByIDAndChildID(db *gorm.DB, session *entity.LearningSession, sessionID, childID string) error {
	return db.Where("session_id = ? AND child_id = ?", sessionID, childID).First(session).Error
}

func (r *StoryPlayRepository) CreateDecision(db *gorm.DB, decision *entity.Decision) error {
	return db.Create(decision).Error
}

func (r *StoryPlayRepository) CompleteLearningSession(db *gorm.DB, session *entity.LearningSession) error {
	return db.Model(session).Updates(map[string]interface{}{
		"completed_at": session.CompletedAt,
	}).Error
}

func (r *StoryPlayRepository) CountDecisionsBySessionID(db *gorm.DB, sessionID string) (int64, error) {
	var count int64
	if err := db.Model(&entity.Decision{}).Where("session_id = ?", sessionID).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *StoryPlayRepository) CreateStorySummary(db *gorm.DB, summary *entity.StorySummary) error {
	return db.Create(summary).Error
}

func (r *StoryPlayRepository) UpdateLearningSessionSummaryID(db *gorm.DB, sessionID, summaryID string) error {
	return db.Model(&entity.LearningSession{}).
		Where("session_id = ?", sessionID).
		Update("summary_id", summaryID).Error
}
