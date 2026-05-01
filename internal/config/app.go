package config

import (
	"example.com/tomo/internal/delivery/http"
	"example.com/tomo/internal/delivery/http/routing"
	"example.com/tomo/internal/helper"
	"example.com/tomo/internal/repository"
	"example.com/tomo/internal/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BootStrapConfig struct {
	DB       *gorm.DB
	App      *echo.Echo
	Log      *zap.Logger
	Validate *validator.Validate
	Config   *viper.Viper
	Redis    *redis.Client
	JWT      *helper.JWTHelper
}

func BootStrap(config *BootStrapConfig) {
	userRepository := repository.NewUserRepository(config.Log)
	childrenRepository := repository.NewChildrenRepository(config.Log)
	coinRepository := repository.NewCoinRepository(config.Log)
	storyHeaderRepository := repository.NewStoryHeaderRepository(config.Log)
	MarketRepository := repository.NewMarketRepository(config.Log)
	userUseCase := usecase.NewUserUsecase(config.DB, config.Log, config.Validate, userRepository, config.JWT, config.Redis)
	childrenUseCase := usecase.NewChildrenUseCase(config.DB, config.Log, config.Validate, childrenRepository, coinRepository, config.JWT, config.Redis)
	storyHeaderUseCase := usecase.NewStoryHeaderUseCase(config.DB, config.Log, config.Validate, storyHeaderRepository)
	MarketUseCase := usecase.NewMarketUseCase(config.DB, config.Log, config.Validate, MarketRepository)
	userController := http.NewUserController(userUseCase, config.Log)
	childrenController := http.NewChildrenController(childrenUseCase, config.Log)
	storyHeaderController := http.NewStoryHeaderController(storyHeaderUseCase, config.Log)
	MarketController := http.NewMarketController(MarketUseCase, config.Log)

	routeConfig := routing.RouteConfig{
		App:                   config.App,
		UserController:        userController,
		ChildrenController:    childrenController,
		StoryHeaderController: storyHeaderController,
		MarketController:      MarketController,
		JWTHelper:             config.JWT,
	}
	routeConfig.SetUp()
}
