package usecase

import (
	"context"
	"net/http"

	"example.com/tomo/internal/model"
	"example.com/tomo/internal/model/converter"
	"example.com/tomo/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StoryHeaderUseCase struct {
	DB                    *gorm.DB
	Log                   *zap.Logger
	Validate              *validator.Validate
	StoryHeaderRepository *repository.StoryHeaderRepository
}

func NewStoryHeaderUseCase(db *gorm.DB, logger *zap.Logger, validate *validator.Validate, storyHeaderRepository *repository.StoryHeaderRepository) *StoryHeaderUseCase {
	return &StoryHeaderUseCase{
		DB:                    db,
		Log:                   logger,
		Validate:              validate,
		StoryHeaderRepository: storyHeaderRepository,
	}
}

func (s *StoryHeaderUseCase) GetAllStoryByParentId(ctx context.Context, parentId string) (*[]model.StoryHeaderResponse, error) {
	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	storyHeader, err := s.StoryHeaderRepository.GetAllStoryHeaderByParentId(tx, parentId)
	if err != nil {
		s.Log.Error("Failed to get story header by parent id", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())

	}
	if err := tx.Commit().Error; err != nil {
		s.Log.Error("Failed to commit transaction")
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	response := make([]model.StoryHeaderResponse, len(*storyHeader))
	for i, storyHeaders := range *storyHeader {
		response[i] = *converter.StoryHeaderToResponse(&storyHeaders)
	}
	return &response, nil
}
