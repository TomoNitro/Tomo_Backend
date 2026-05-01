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
