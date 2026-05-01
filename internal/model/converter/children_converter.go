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

func ChildrenListToResponse(children []entity.Children) []model.ChildrenListResponse {
	responses := make([]model.ChildrenListResponse, 0, len(children))

	for _, child := range children {
		responses = append(responses, model.ChildrenListResponse{
			ID:        child.ID,
			Name:      child.Name,
			CreatedAt: child.CreatedAt,
		})
	}

	return responses
}

func ChildrenDeleteToResponse(child *entity.Children) *model.ChildrenDeleteResponse {
	return &model.ChildrenDeleteResponse{
		ID:   child.ID,
		Name: child.Name,
	}
}
