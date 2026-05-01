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
func (c *ChildrenRepository) FindByID(db *gorm.DB, child *entity.Children, id string) error {
	return db.Where("id = ?", id).First(child).Error
}

func (c *ChildrenRepository) FindByParentID(db *gorm.DB, parentID string) ([]entity.Children, error) {
	var children []entity.Children

	err := db.Where("parent_id = ?", parentID).Order("created_at ASC").Find(&children).Error
	if err != nil {
		return nil, err
	}

	return children, nil
}
