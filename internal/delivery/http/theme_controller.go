package http

import (
	"net/http"

	"example.com/tomo/internal/model"
	"example.com/tomo/internal/usecase"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type ThemeController struct {
	ThemeUseCase *usecase.ThemeUseCase
	Logger       *zap.Logger
}

func NewThemeController(usecase *usecase.ThemeUseCase, logger *zap.Logger) *ThemeController {
	return &ThemeController{
		ThemeUseCase: usecase,
		Logger:       logger,
	}
}

func (t *ThemeController) GetThemes(ctx *echo.Context) error {
	response, err := t.ThemeUseCase.GetFinanceAndStoryThemes(ctx.Request().Context())
	if err != nil {
		t.Logger.Error("Failed to get themes", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.StoryFinaceThemeResponse]{
		Message: "Success get themes",
		Data:    response,
	})
}
