package converter

import (
	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/model"
)

func UserRegisterToResponse(user *entity.User, accessToken, refreshToken string) *model.UserRegisterResponse {
	return &model.UserRegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		Token: model.UserLoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
}

func UserLoginToResponse(user *entity.User, accessToken, refreshToken string) *model.UserLoginResponse {
	return &model.UserLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
func RefreshTokenToReponse(accessToken string) *model.ResponseRefreshToken {
	return &model.ResponseRefreshToken{
		AccessToken: accessToken,
	}
}
