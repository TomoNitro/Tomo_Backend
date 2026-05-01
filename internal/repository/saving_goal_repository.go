package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SavingGoalRepository struct {
	Repository[entity.SavingGoal]
	Log *zap.Logger
}

func NewSavingGoalRepository(log *zap.Logger) *SavingGoalRepository {
	return &SavingGoalRepository{
		Log: log,
	}
}

func (s *SavingGoalRepository) FindByChildID(db *gorm.DB, goal *entity.SavingGoal, childID string) error {
	return db.Where("child_id = ?", childID).First(goal).Error
}
