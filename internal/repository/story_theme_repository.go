package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
)

type StoryThemeRepository struct {
	Repository[entity.StoryTheme]
	Log *zap.Logger
}

func NewStoryThemeRepository(log *zap.Logger) *StoryThemeRepository {
	return &StoryThemeRepository{
		Log: log,
	}
}
