package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ChildBadgeRepository struct {
	Repository[entity.ChildBadge]
	Log *zap.Logger
}

func NewChildBadgeRepository(log *zap.Logger) *ChildBadgeRepository {
	return &ChildBadgeRepository{Log: log}
}

func (r *ChildBadgeRepository) FindByChildID(db *gorm.DB, childID string) ([]entity.ChildBadge, error) {
	var badges []entity.ChildBadge
	if err := db.Where("child_id = ?", childID).Find(&badges).Error; err != nil {
		return nil, err
	}
	return badges, nil
}

func (r *ChildBadgeRepository) FindByChildAndBadge(db *gorm.DB, childBadge *entity.ChildBadge, childID, badgeID string) error {
	return db.Where("child_id = ? AND badge_id = ?", childID, badgeID).First(childBadge).Error
}
