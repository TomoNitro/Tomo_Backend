package http

import (
	"net/http"

	"example.com/tomo/internal/helper"
	"example.com/tomo/internal/model"
	"example.com/tomo/internal/usecase"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type ChildrenController struct {
	ChildrenUseCase *usecase.ChildrenUseCase
	Logger          *zap.Logger
}

func NewChildrenController(usecase *usecase.ChildrenUseCase, logger *zap.Logger) *ChildrenController {
	return &ChildrenController{
		ChildrenUseCase: usecase,
		Logger:          logger,
	}
}

func (c *ChildrenController) ChildrenRegister(ctx *echo.Context) error {
	request := new(model.ChildrenRequest)
	if err := ctx.Bind(request); err != nil {
		c.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	parentID := helper.GetActorID(ctx)
	response, err := c.ChildrenUseCase.ChildrenRegister(ctx.Request().Context(), parentID, request)
	if err != nil {
		c.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, model.WebResponse[*model.ChildrenRegisterResponse]{Message: "Success create child", Data: response})
}

func (c *ChildrenController) GetChildrenByParent(ctx *echo.Context) error {
	parentID := helper.GetActorID(ctx)

	response, err := c.ChildrenUseCase.GetChildrenByParent(ctx.Request().Context(), parentID)
	if err != nil {
		c.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[[]model.ChildrenListResponse]{Message: "Success get children", Data: response})
}

func (c *ChildrenController) DeleteChildrenByParent(ctx *echo.Context) error {
	parentID := helper.GetActorID(ctx)
	childID := ctx.Param("id")

	response, err := c.ChildrenUseCase.DeleteChildrenByParent(ctx.Request().Context(), parentID, childID)
	if err != nil {
		c.Logger.Error(err.Error())
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.ChildrenDeleteResponse]{Message: "Success delete child", Data: response})
}

func (c *ChildrenController) ChildrenLogin(ctx *echo.Context) error {
	request := new(model.ChildrenLoginRequest)
	if err := ctx.Bind(request); err != nil {
		c.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	response, err := c.ChildrenUseCase.ChildrenLogin(ctx.Request().Context(), request)
	if err != nil {
		c.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, model.WebResponse[*model.ChildrenLoginResponse]{Message: "Success login child", Data: response})
}

func (c *ChildrenController) SetSavingGoal(ctx *echo.Context) error {
	childID := helper.GetActorID(ctx)
	if childID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	marketID := ctx.Param("id")
	if marketID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "market id is required")
	}

	response, err := c.ChildrenUseCase.SetSavingGoal(ctx.Request().Context(), childID, marketID)
	if err != nil {
		c.Logger.Error("Failed to set saving goal", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.SavingGoalResponse]{Message: "Success set saving goal", Data: response})
}

func (c *ChildrenController) GetChildrenCoin(ctx *echo.Context) error {
	childID := helper.GetActorID(ctx)
	if childID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	response, err := c.ChildrenUseCase.GetChildrenCoin(ctx.Request().Context(), childID)
	if err != nil {
		c.Logger.Error("Failed to get child coin", zap.Error(err))
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.ChildrenCoinResponse]{Message: "Success get child coin", Data: response})
}
