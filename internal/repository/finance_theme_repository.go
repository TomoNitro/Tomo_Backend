package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
)

type FinanceThemeRepository struct {
	Repository[entity.FinanceTheme]
	Log *zap.Logger
}

func NewFinanceThemeRepository(log *zap.Logger) *FinanceThemeRepository {
	return &FinanceThemeRepository{
		Log: log,
	}
}
