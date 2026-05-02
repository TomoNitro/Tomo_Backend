package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BadgeRepository struct {
	Repository[entity.Badge]
	Log *zap.Logger
}

func NewBadgeRepository(log *zap.Logger) *BadgeRepository {
	return &BadgeRepository{Log: log}
}

func (r *BadgeRepository) FindAll(db *gorm.DB) ([]entity.Badge, error) {
	var badges []entity.Badge
	if err := db.Order("level_required ASC").Find(&badges).Error; err != nil {
		return nil, err
	}
	return badges, nil
}
