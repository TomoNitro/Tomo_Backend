package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ExpTransactionRepository struct {
	Repository[entity.ExpTransaction]
	Log *zap.Logger
}

func NewExpTransactionRepository(log *zap.Logger) *ExpTransactionRepository {
	return &ExpTransactionRepository{
		Log: log,
	}
}

func (r *ExpTransactionRepository) CountBySourceAndReferenceID(db *gorm.DB, source, referenceID string) (int64, error) {
	var count int64
	if err := db.Model(&entity.ExpTransaction{}).
		Where("source = ? AND reference_id = ?", source, referenceID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *ExpTransactionRepository) FindBySourceAndReferenceID(db *gorm.DB, transaction *entity.ExpTransaction, source, referenceID string) error {
	return db.Where("source = ? AND reference_id = ?", source, referenceID).First(transaction).Error
}
