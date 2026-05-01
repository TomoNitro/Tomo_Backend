package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository struct {
	Repository[entity.User]
	Log *zap.Logger
}

func NewUserRepository(Log *zap.Logger) *UserRepository {
	return &UserRepository{
		Log: Log,
	}
}
func (u *UserRepository) Login(db *gorm.DB, user *entity.User, email string) error {
	return db.Where("email = ?", email).First(user).Error
}
