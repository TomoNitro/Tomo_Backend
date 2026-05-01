package routing

import (
	"github.com/labstack/echo/v5"
)

type RouteConfig struct {
	App *echo.Echo
}

func (r *RouteConfig) SetUp() {
	r.SetupGuestRoute()
}
func (r *RouteConfig) SetupGuestRoute() {

}
