package converter

import (
	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/model"
)

func ChildrenRegisterToResponse(child *entity.Children, accessToken, refreshToken string) *model.ChildrenRegisterResponse {
	return &model.ChildrenRegisterResponse{
		ID:        child.ID,
		Name:      child.Name,
		CreatedAt: child.CreatedAt,
		Token: model.ChildrenLoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
}
func ChildrenLoginToResponse(child *entity.Children, accessToken, refreshToken string) *model.ChildrenLoginResponse {
	return &model.ChildrenLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
