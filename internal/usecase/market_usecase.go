package usecase

import (
	"context"
	"net/http"

	"example.com/tomo/internal/model"
	"example.com/tomo/internal/model/converter"
	"example.com/tomo/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MarketUseCase struct {
	DB               *gorm.DB
	Log              *zap.Logger
	Validate         *validator.Validate
	MarketRepository *repository.MarketRepository
}

func NewMarketUseCase(db *gorm.DB, logger *zap.Logger, validate *validator.Validate, marketRepository *repository.MarketRepository) *MarketUseCase {
	return &MarketUseCase{
		DB:               db,
		Log:              logger,
		Validate:         validate,
		MarketRepository: marketRepository,
	}
}
func (p *MarketUseCase) GetAllMarkets(ctx context.Context) (*[]model.MarketResponse, error) {
	tx := p.DB.WithContext(ctx).Begin()
	defer tx.Rollback()
	markets, err := p.MarketRepository.Find(tx)
	if err != nil {
		p.Log.Error("Failed to get markets")
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := tx.Commit().Error; err != nil {
		p.Log.Error("Failed to commit transaction")
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	response := make([]model.MarketResponse, len(*markets))
	for i, market := range *markets {
		response[i] = *converter.ToMarketResponse(&market)
	}
	return &response, nil
}
