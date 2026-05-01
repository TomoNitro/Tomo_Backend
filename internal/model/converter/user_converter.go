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
func UserUpdateToResponse(user *entity.User) *model.UserUpdateResponse {
	return &model.UserUpdateResponse{
		Username: user.Username,
		Email:    user.Email,
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

func ParentInfoToResponse(user *entity.User) *model.ParentInfoResponse {
	return &model.ParentInfoResponse{
		Username: user.Username,
		Email:    user.Email,
	}
}
