package converter

import (
	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/model"
)

func ToMarketResponse(market *entity.Market) *model.MarketResponse {
	return &model.MarketResponse{
		ID:        market.ID,
		Title:     market.Title,
		ImageURL:  market.ImageURL,
		Price:     market.Price,
		CreatedAt: market.CreatedAt,
	}
}
