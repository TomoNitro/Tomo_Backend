package model

import "time"

type MarketResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	ImageURL  string    `json:"image_url"`
	Price     int       `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}
