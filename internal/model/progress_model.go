package model

import "time"

type BadgeResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	LevelRequired int        `json:"level_required"`
	ImageURL      *string    `json:"image_url,omitempty"`
	Earned        bool       `json:"earned"`
	EarnedAt      *time.Time `json:"earned_at,omitempty"`
}

type ChildProgressResponse struct {
	TotalExp       int             `json:"total_exp"`
	Level          int             `json:"level"`
	NextLevelExp   int             `json:"next_level_exp"`
	ExpToNextLevel int             `json:"exp_to_next_level"`
	Badges         []BadgeResponse `json:"badges"`
}
