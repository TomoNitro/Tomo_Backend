package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ChildrenRepository struct {
	Repository[entity.Children]
	Log *zap.Logger
}

func NewChildrenRepository(Log *zap.Logger) *ChildrenRepository {
	return &ChildrenRepository{
		Log: Log,
	}
}
func (c *ChildrenRepository) Login(db *gorm.DB, child *entity.Children, name string) error {
	return db.Where("name = ?", name).First(child).Error
}
