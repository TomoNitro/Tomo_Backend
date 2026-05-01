package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CoinRepository struct {
	Repository[entity.CoinTransaction]
	Log *zap.Logger
}

func NewCoinRepository(log *zap.Logger) *CoinRepository {
	return &CoinRepository{
		Log: log,
	}
}

func (c *CoinRepository) FindByChildID(db *gorm.DB, coin *entity.CoinTransaction, childID string) error {
	return db.Where("child_id = ?", childID).First(coin).Error
}
