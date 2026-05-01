package usecase

import (
	"context"
	"net/http"

	"example.com/tomo/internal/model"
	"example.com/tomo/internal/repository"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ThemeUseCase struct {
	DB                     *gorm.DB
	Log                    *zap.Logger
	FinanceThemeRepository *repository.FinanceThemeRepository
	StoryThemeRepository   *repository.StoryThemeRepository
}

func NewThemeUseCase(db *gorm.DB, log *zap.Logger, financeThemeRepository *repository.FinanceThemeRepository, storyThemeRepository *repository.StoryThemeRepository) *ThemeUseCase {
	return &ThemeUseCase{
		DB:                     db,
		Log:                    log,
		FinanceThemeRepository: financeThemeRepository,
		StoryThemeRepository:   storyThemeRepository,
	}
}

func (t *ThemeUseCase) GetFinanceAndStoryThemes(ctx context.Context) (*model.StoryFinaceThemeResponse, error) {
	tx := t.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	financeThemes, err := t.FinanceThemeRepository.Find(tx)
	if err != nil {
		t.Log.Error("Failed to get finance themes", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	storyThemes, err := t.StoryThemeRepository.Find(tx)
	if err != nil {
		t.Log.Error("Failed to get story themes", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		t.Log.Error("Failed to commit transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	financeResponse := make([]model.Finance, len(*financeThemes))
	for i, theme := range *financeThemes {
		financeResponse[i] = model.Finance{Topic: theme.Topic}
	}

	storyResponse := make([]model.Story, len(*storyThemes))
	for i, theme := range *storyThemes {
		storyResponse[i] = model.Story{Topic: theme.Topic, FullStory: theme.FullStory}
	}

	return &model.StoryFinaceThemeResponse{
		Finance: financeResponse,
		Story:   storyResponse,
	}, nil
}
