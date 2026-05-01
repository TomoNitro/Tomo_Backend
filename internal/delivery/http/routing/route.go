package routing

import (
	"example.com/tomo/internal/delivery/http"
	"example.com/tomo/internal/helper"
	"github.com/labstack/echo/v5"
)

type RouteConfig struct {
	App            *echo.Echo
	UserController *http.UserController
	JWTHelper      *helper.JWTHelper
}

func (r *RouteConfig) SetUp() {
	r.SetupGuestRoute()
}
func (r *RouteConfig) SetupGuestRoute() {
	user := r.App.Group("/api/user")
	user.POST("/register", r.UserController.Register)
	user.POST("/login", r.UserController.Login)
	user.POST("/refresh-token", r.UserController.RefreshToken)
}
