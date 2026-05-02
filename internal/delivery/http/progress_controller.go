package http

import (
	"net/http"

	"example.com/tomo/internal/helper"
	"example.com/tomo/internal/model"
	"example.com/tomo/internal/usecase"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type ProgressController struct {
	ProgressUseCase *usecase.ProgressUseCase
	Logger          *zap.Logger
}

func NewProgressController(usecase *usecase.ProgressUseCase, logger *zap.Logger) *ProgressController {
	return &ProgressController{
		ProgressUseCase: usecase,
		Logger:          logger,
	}
}

func (c *ProgressController) GetChildProgress(ctx *echo.Context) error {
	childID := helper.GetActorID(ctx)
	if childID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	response, err := c.ProgressUseCase.GetChildProgress(ctx.Request().Context(), childID)
	if err != nil {
		c.Logger.Error("failed to get child progress", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.ChildProgressResponse]{
		Message: "Success get child progress",
		Data:    response,
	})
}
