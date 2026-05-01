package routing

import (
	"example.com/tomo/internal/delivery/http"
	"example.com/tomo/internal/delivery/http/middleware"
	"example.com/tomo/internal/helper"
	"github.com/labstack/echo/v5"
)

type RouteConfig struct {
	App                   *echo.Echo
	UserController        *http.UserController
	ChildrenController    *http.ChildrenController
	ThemeController       *http.ThemeController
	StoryHeaderController *http.StoryHeaderController
	StoryPlayController   *http.StoryPlayController
	MarketController      *http.MarketController
	JWTHelper             *helper.JWTHelper
}

func (r *RouteConfig) SetUp() {
	r.SetupGuestRoute()
}
func (r *RouteConfig) SetupGuestRoute() {
	theme := r.App.Group("/api/themes")
	theme.GET("", r.ThemeController.GetThemes)

	user := r.App.Group("/api/user")
	user.POST("/register", r.UserController.Register)
	user.POST("/login", r.UserController.Login)
	user.POST("/refresh-token", r.UserController.RefreshToken)

	parentOnly := r.App.Group("/api/parent", middleware.AuthMiddleware(r.JWTHelper), middleware.ParentOnly())
	parentOnly.GET("/story-headers", r.StoryHeaderController.GetAllStoryByParentId)
	parentOnly.POST("/story-headers/generate", r.StoryHeaderController.CreateStory)
	parentOnly.PUT("/update", r.UserController.UpdateProfile)
	parentOnly.GET("/info", r.UserController.GetParentInfo)

	childrenOnly := r.App.Group("/api/children", middleware.AuthMiddleware(r.JWTHelper), middleware.ChildOnly())
	childrenOnly.GET("/markets", r.MarketController.GetAllMarket)
	childrenOnly.GET("/coins", r.ChildrenController.GetChildrenCoin)
	childrenOnly.POST("/saving-goal/:id", r.ChildrenController.SetSavingGoal)
	childrenOnly.PUT("/name", r.ChildrenController.UpdateChildName)
	childrenOnly.GET("/story-headers", r.StoryHeaderController.GetAllStoryByChildId)
	childrenOnly.POST("/stories/:storyId/start", r.StoryPlayController.StartStory)

	sessionsOnly := r.App.Group("/api/sessions", middleware.AuthMiddleware(r.JWTHelper), middleware.ChildOnly())
	sessionsOnly.POST("/:sessionId/decision", r.StoryPlayController.MakeDecision)

	children := r.App.Group("/api/children")
	children.POST("/login", r.ChildrenController.ChildrenLogin)

	parentChildren := r.App.Group("/api/children", middleware.AuthMiddleware(r.JWTHelper), middleware.ParentOnly())
	parentChildren.GET("", r.ChildrenController.GetChildrenByParent)
	parentChildren.DELETE("/:id", r.ChildrenController.DeleteChildrenByParent)
	parentChildren.POST("/register", r.ChildrenController.ChildrenRegister)

}
