package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StoryHeaderRepository struct {
	Repository[entity.StoryHeader]
	Log *zap.Logger
}

func NewStoryHeaderRepository(Log *zap.Logger) *StoryHeaderRepository {
	return &StoryHeaderRepository{
		Log: Log,
	}
}
func (r *Repository[T]) GetAllStoryHeaderByParentId(db *gorm.DB, parentId string) (*[]entity.StoryHeader, error) {
	var storyHeader []entity.StoryHeader

	if err := db.Where("parent_id = ?", parentId).Find(&storyHeader).Error; err != nil {
		return nil, err
	}
	return &storyHeader, nil
}

func (r *StoryHeaderRepository) GetAllStoryHeaderByParentAndChildId(db *gorm.DB, parentId, childId string) (*[]entity.StoryHeader, error) {
	var storyHeader []entity.StoryHeader

	subQuery := db.Model(&entity.LearningSession{}).
		Select("story_id").
		Where("child_id = ? AND completed_at IS NOT NULL", childId)

	if err := db.Where("parent_id = ?", parentId).
		Where("story_id NOT IN (?)", subQuery).
		Find(&storyHeader).Error; err != nil {
		return nil, err
	}

	return &storyHeader, nil
}
