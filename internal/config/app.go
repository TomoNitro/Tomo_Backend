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
	dashboardRepository := repository.NewDashboardRepository(config.Log)
	coinRepository := repository.NewCoinRepository(config.Log)
	savingGoalRepository := repository.NewSavingGoalRepository(config.Log)
	financeThemeRepository := repository.NewFinanceThemeRepository(config.Log)
	storyThemeRepository := repository.NewStoryThemeRepository(config.Log)
	storyHeaderRepository := repository.NewStoryHeaderRepository(config.Log)
	storyPlayRepository := repository.NewStoryPlayRepository(config.Log)
	childProgressRepository := repository.NewChildProgressRepository(config.Log)
	badgeRepository := repository.NewBadgeRepository(config.Log)
	childBadgeRepository := repository.NewChildBadgeRepository(config.Log)
	expTransactionRepository := repository.NewExpTransactionRepository(config.Log)
	MarketRepository := repository.NewMarketRepository(config.Log)
	storyWebhookURL := config.Config.GetString("STORY_WEBHOOK_URL")
	if storyWebhookURL == "" {
		storyWebhookURL = "https://williamdarma-n8n.hf.space/webhook/create-story"
	}
	summaryWebhookURL := config.Config.GetString("STORY_SUMMARY_WEBHOOK_URL")
	if summaryWebhookURL == "" {
		summaryWebhookURL = "https://williamdarma-n8n.hf.space/webhook/generate-story-summary"
	}
	nodeAudioWebhookURL := config.Config.GetString("STORY_NODE_AUDIO_WEBHOOK_URL")
	if nodeAudioWebhookURL == "" {
		nodeAudioWebhookURL = "https://williamdarma-n8n.hf.space/webhook/generate-image-audio"
	}
	dashboardSummaryWebhookURL := config.Config.GetString("DASHBOARD_SUMMARY_WEBHOOK_URL")
	if dashboardSummaryWebhookURL == "" {
		dashboardSummaryWebhookURL = "https://williamdarma-n8n.hf.space/webhook/generate-dashboard-summary"
	}
	userUseCase := usecase.NewUserUsecase(config.DB, config.Log, config.Validate, userRepository, config.JWT, config.Redis)
	childrenUseCase := usecase.NewChildrenUseCase(config.DB, config.Log, config.Validate, childrenRepository, coinRepository, savingGoalRepository, MarketRepository, config.JWT, config.Redis)
	dashboardUseCase := usecase.NewDashboardUseCase(config.DB, config.Log, childrenRepository, dashboardRepository, dashboardSummaryWebhookURL)
	themeUseCase := usecase.NewThemeUseCase(config.DB, config.Log, financeThemeRepository, storyThemeRepository)
	storyHeaderUseCase := usecase.NewStoryHeaderUseCase(config.DB, config.Log, config.Validate, storyHeaderRepository, storyWebhookURL)
	storyPlayUseCase := usecase.NewStoryPlayUseCase(config.DB, config.Log, config.Validate, storyPlayRepository, childrenRepository, childProgressRepository, expTransactionRepository, coinRepository, summaryWebhookURL, nodeAudioWebhookURL)
	progressUseCase := usecase.NewProgressUseCase(config.DB, config.Log, childProgressRepository, badgeRepository, childBadgeRepository)
	MarketUseCase := usecase.NewMarketUseCase(config.DB, config.Log, config.Validate, MarketRepository)
	userController := http.NewUserController(userUseCase, config.Log)
	childrenController := http.NewChildrenController(childrenUseCase, config.Log)
	dashboardController := http.NewDashboardController(dashboardUseCase, config.Log)
	themeController := http.NewThemeController(themeUseCase, config.Log)
	storyHeaderController := http.NewStoryHeaderController(storyHeaderUseCase, config.Log)
	storyPlayController := http.NewStoryPlayController(storyPlayUseCase, config.Log)
	progressController := http.NewProgressController(progressUseCase, config.Log)
	MarketController := http.NewMarketController(MarketUseCase, config.Log)

	routeConfig := routing.RouteConfig{
		App:                   config.App,
		UserController:        userController,
		ChildrenController:    childrenController,
		DashboardController:   dashboardController,
		ThemeController:       themeController,
		StoryHeaderController: storyHeaderController,
		StoryPlayController:   storyPlayController,
		ProgressController:    progressController,
		MarketController:      MarketController,
		JWTHelper:             config.JWT,
	}
	routeConfig.SetUp()
}
