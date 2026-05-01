package http

import (
	"net/http"

	"example.com/tomo/internal/model"
	"example.com/tomo/internal/usecase"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type UserController struct {
	UserUseCase *usecase.UserUseCase
	Logger      *zap.Logger
}

func NewUserController(usecase *usecase.UserUseCase, logger *zap.Logger) *UserController {
	return &UserController{
		UserUseCase: usecase,
		Logger:      logger,
	}
}
func (u *UserController) Register(ctx *echo.Context) error {
	request := new(model.UserRequest)

	if err := ctx.Bind(request); err != nil {
		u.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	response, err := u.UserUseCase.UserRegister(ctx.Request().Context(), request)
	if err != nil {
		u.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, model.WebResponse[*model.UserRegisterResponse]{Message: "User Register Successful", Data: response})
}
func (u *UserController) Login(ctx *echo.Context) error {
	request := new(model.UserLoginRequest)

	if err := ctx.Bind(request); err != nil {
		u.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	response, err := u.UserUseCase.UserLogin(ctx.Request().Context(), request)
	if err != nil {
		u.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, model.WebResponse[*model.UserLoginResponse]{Message: "User Login Success", Data: response})
}
func (u *UserController) RefreshToken(ctx *echo.Context) error {
	request := new(model.RequestRefreshToken)
	if err := ctx.Bind(request); err != nil {
		u.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	response, err := u.UserUseCase.RefreshToken(ctx.Request().Context(), request)
	if err != nil {
		u.Logger.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, model.WebResponse[*model.ResponseRefreshToken]{Message: "Success refresh token", Data: response})
}
