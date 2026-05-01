package repository

import (
	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
)

type MarketRepository struct {
	Repository[entity.Market]
	Log *zap.Logger
}

func NewMarketRepository(Log *zap.Logger) *MarketRepository {
	return &MarketRepository{
		Log: Log,
	}
}
