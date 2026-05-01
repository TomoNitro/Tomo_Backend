package http

import (
	"net/http"

	"example.com/tomo/internal/helper"
	"example.com/tomo/internal/model"
	"example.com/tomo/internal/usecase"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type StoryHeaderController struct {
	StoryHeaderUseCase *usecase.StoryHeaderUseCase
	Logger             *zap.Logger
}

func NewStoryHeaderController(usecase *usecase.StoryHeaderUseCase, logger *zap.Logger) *StoryHeaderController {
	return &StoryHeaderController{
		StoryHeaderUseCase: usecase,
		Logger:             logger,
	}
}
func (p *StoryHeaderController) GetAllStoryByParentId(ctx *echo.Context) error {
	actorType := helper.GetActorType(ctx)
	if actorType != "parent" {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}

	parentID := helper.GetActorID(ctx)
	response, err := p.StoryHeaderUseCase.GetAllStoryByParentId(ctx.Request().Context(), parentID)
	if err != nil {
		p.Logger.Error("Failed to get story headers")
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}
	return ctx.JSON(http.StatusOK, model.WebResponse[*[]model.StoryHeaderResponse]{Data: response})
}
func (p *StoryHeaderController) GetAllStoryByChildId(ctx *echo.Context) error {
	actorType := helper.GetActorType(ctx)
	if actorType != "child" {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}

	parentID := helper.GetParentID(ctx)
	response, err := p.StoryHeaderUseCase.GetAllStoryByParentId(ctx.Request().Context(), parentID)
	if err != nil {
		p.Logger.Error("Failed to get story headers")
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}
	return ctx.JSON(http.StatusOK, model.WebResponse[*[]model.StoryHeaderResponse]{Data: response})
}
