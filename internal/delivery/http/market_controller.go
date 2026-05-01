package http

import (
	"net/http"

	"example.com/tomo/internal/model"
	"example.com/tomo/internal/usecase"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type MarketController struct {
	MarketUseCase *usecase.MarketUseCase
	Logger        *zap.Logger
}

func NewMarketController(usecase *usecase.MarketUseCase, logger *zap.Logger) *MarketController {
	return &MarketController{
		MarketUseCase: usecase,
		Logger:        logger,
	}
}
func (p *MarketController) GetAllMarket(ctx *echo.Context) error {
	response, err := p.MarketUseCase.GetAllMarkets(ctx.Request().Context())
	if err != nil {
		p.Logger.Error("Failed to get product")
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}
	return ctx.JSON(http.StatusOK, model.WebResponse[*[]model.MarketResponse]{Data: response})
}
