package model

import "time"

type ChildrenRequest struct {
	Name string `json:"name" validate:"required,max=50"`
	Pin  string `json:"pin" validate:"required,max=50"`
}
type ChildrenLoginRequest struct {
	ChildID string `json:"childId" validate:"required,max=50"`
	Pin     string `json:"pin" validate:"required,max=50"`
}
type ChildrenListResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
type ChildrenRegisterResponse struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	CreatedAt time.Time             `json:"created_at"`
	Token     ChildrenLoginResponse `json:"Token"`
}
type ChildrenLoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
