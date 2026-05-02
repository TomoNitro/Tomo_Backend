package http

import (
	"net/http"

	"example.com/tomo/internal/helper"
	"example.com/tomo/internal/model"
	"example.com/tomo/internal/usecase"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type DashboardController struct {
	DashboardUseCase *usecase.DashboardUseCase
	Logger           *zap.Logger
}

func NewDashboardController(usecase *usecase.DashboardUseCase, logger *zap.Logger) *DashboardController {
	return &DashboardController{
		DashboardUseCase: usecase,
		Logger:           logger,
	}
}

func (c *DashboardController) GetChildDashboard(ctx *echo.Context) error {
	parentID := helper.GetActorID(ctx)
	if parentID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	childID := ctx.Param("childId")
	response, err := c.DashboardUseCase.GetChildDashboard(ctx.Request().Context(), parentID, childID)
	if err != nil {
		c.Logger.Error("failed to get child dashboard", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.ParentChildDashboardResponse]{
		Message: "Success get child dashboard",
		Data:    response,
	})
}

func (c *DashboardController) GenerateChildDashboardSummary(ctx *echo.Context) error {
	parentID := helper.GetActorID(ctx)
	if parentID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	childID := ctx.Param("childId")
	response, err := c.DashboardUseCase.GenerateChildDashboardSummary(ctx.Request().Context(), parentID, childID)
	if err != nil {
		c.Logger.Error("failed to generate child dashboard summary", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.DashboardSummaryResponse]{
		Message: "Success generate child dashboard summary",
		Data:    response,
	})
}

func (c *DashboardController) GetLatestChildDashboardSummary(ctx *echo.Context) error {
	parentID := helper.GetActorID(ctx)
	if parentID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	childID := ctx.Param("childId")
	response, err := c.DashboardUseCase.GetLatestChildDashboardSummary(ctx.Request().Context(), parentID, childID)
	if err != nil {
		c.Logger.Error("failed to get latest child dashboard summary", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.DashboardSummaryResponse]{
		Message: "Success get child dashboard summary",
		Data:    response,
	})
}
