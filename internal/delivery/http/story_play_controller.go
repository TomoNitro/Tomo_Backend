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

func (c *StoryPlayController) GenerateStoryNodeAudio(ctx *echo.Context) error {
	nodeID := ctx.Param("nodeId")
	actorChildID := helper.GetActorID(ctx)

	response, err := c.StoryPlayUseCase.GenerateStoryNodeAudio(ctx.Request().Context(), actorChildID, nodeID)
	if err != nil {
		c.Logger.Error("failed to generate story node audio", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.StoryNodeAudioResponse]{
		Message: "Success generate story node audio",
		Data:    response,
	})
}

func (c *StoryPlayController) GenerateStorySummary(ctx *echo.Context) error {
	sessionID := ctx.Param("sessionId")
	actorChildID := helper.GetActorID(ctx)

	response, err := c.StoryPlayUseCase.GenerateStorySummary(ctx.Request().Context(), actorChildID, sessionID)
	if err != nil {
		c.Logger.Error("failed to generate story summary", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.StorySummaryRewardResponse]{
		Message: "Success generate story summary",
		Data:    response,
	})
}
