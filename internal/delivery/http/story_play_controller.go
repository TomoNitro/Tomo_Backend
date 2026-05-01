package http

import (
	"net/http"

	"example.com/tomo/internal/helper"
	"example.com/tomo/internal/model"
	"example.com/tomo/internal/usecase"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type StoryPlayController struct {
	StoryPlayUseCase *usecase.StoryPlayUseCase
	Logger           *zap.Logger
}

func NewStoryPlayController(storyPlayUseCase *usecase.StoryPlayUseCase, logger *zap.Logger) *StoryPlayController {
	return &StoryPlayController{
		StoryPlayUseCase: storyPlayUseCase,
		Logger:           logger,
	}
}

func (c *StoryPlayController) StartStory(ctx *echo.Context) error {
	actorChildID := helper.GetActorID(ctx)
	storyID := ctx.Param("storyId")

	response, err := c.StoryPlayUseCase.StartStory(ctx.Request().Context(), actorChildID, storyID)
	if err != nil {
		c.Logger.Error("failed to start story", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.StoryPlayStartResponse]{
		Message: "Success start story",
		Data:    response,
	})
}

func (c *StoryPlayController) MakeDecision(ctx *echo.Context) error {
	request := new(model.StoryDecisionRequest)
	if err := ctx.Bind(request); err != nil {
		c.Logger.Error("failed to bind story decision request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	sessionID := ctx.Param("sessionId")
	actorChildID := helper.GetActorID(ctx)
	response, err := c.StoryPlayUseCase.MakeDecision(ctx.Request().Context(), actorChildID, sessionID, request)
	if err != nil {
		c.Logger.Error("failed to make story decision", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.StoryPlayDecisionResponse]{
		Message: "Success make decision",
		Data:    response,
	})
}
