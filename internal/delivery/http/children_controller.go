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
	parentID := helper.GetUserID(ctx)
	response, err := c.ChildrenUseCase.ChildrenRegister(ctx.Request().Context(), parentID, request)
	if err != nil {
		c.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, model.WebResponse[*model.ChildrenRegisterResponse]{Message: "Success create child", Data: response})
}
func (c *ChildrenController) ChildrenLogin(ctx *echo.Context) error {
	request := new(model.ChildrenRequest)
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
