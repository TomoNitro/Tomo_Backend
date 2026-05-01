package config

import (
	"example.com/tomo/internal/delivery/http/routing"
	"example.com/tomo/internal/helper"
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

	routeConfig := routing.RouteConfig{
		App: config.App,
	}
	routeConfig.SetUp()
}
