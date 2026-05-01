package routing

import (
	"example.com/tomo/internal/delivery/http"
	"example.com/tomo/internal/delivery/http/middleware"
	"example.com/tomo/internal/helper"
	"github.com/labstack/echo/v5"
)

type RouteConfig struct {
	App                *echo.Echo
	UserController     *http.UserController
	ChildrenController *http.ChildrenController
	JWTHelper          *helper.JWTHelper
}

func (r *RouteConfig) SetUp() {
	r.SetupGuestRoute()
}
func (r *RouteConfig) SetupGuestRoute() {
	user := r.App.Group("/api/user")
	user.POST("/register", r.UserController.Register)
	user.POST("/login", r.UserController.Login)
	user.POST("/refresh-token", r.UserController.RefreshToken)

	children := r.App.Group("/api/children")
	children.POST("/login", r.ChildrenController.ChildrenLogin)

	parentChildren := r.App.Group("/api/children", middleware.AuthMiddleware(r.JWTHelper), middleware.ParentOnly())
	parentChildren.POST("/register", r.ChildrenController.ChildrenRegister)
}
