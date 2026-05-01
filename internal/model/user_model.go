package model

import "time"

type UserRequest struct {
	Email    string `json:"email" validate:"required,max=50"`
	Username string `json:"username" validate:"required,max=50"`
	Password string `json:"password" validate:"required,max=50"`
}
type UserUpdateRequest struct {
	Username string `json:"username" validate:"required,max=50"`
	Email    string `json:"email" validate:"required,max=50"`
}
type UserUpdateResponse struct {
	Username string `json:"username" validate:"required,max=50"`
	Email    string `json:"email" validate:"required,max=50"`
}
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,max=50"`
	Password string `json:"password" validate:"required,max=50"`
}
type UserRegisterResponse struct {
	ID        string            `json:"id"`
	Email     string            `json:"email"`
	Username  string            `json:"username"`
	CreatedAt time.Time         `json:"created_at"`
	Token     UserLoginResponse `json:"Token"`
}
type UserLoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
type RequestRefreshToken struct {
	RefreshToken string `json:"refreshToken"`
}
type ResponseRefreshToken struct {
	AccessToken string `json:"accessToken"`
}
