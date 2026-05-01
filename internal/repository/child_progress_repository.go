package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ChildProgressRepository struct {
	Repository[entity.ChildProgress]
	Log *zap.Logger
}

func NewChildProgressRepository(log *zap.Logger) *ChildProgressRepository {
	return &ChildProgressRepository{
		Log: log,
	}
}

func (r *ChildProgressRepository) FindByChildID(db *gorm.DB, progress *entity.ChildProgress, childID string) error {
	return db.Where("child_id = ?", childID).First(progress).Error
}
