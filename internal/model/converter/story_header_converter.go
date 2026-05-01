package converter

import (
	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/model"
)

func StoryHeaderToResponse(storyHeader *entity.StoryHeader) *model.StoryHeaderResponse {
	return &model.StoryHeaderResponse{
		ID:          storyHeader.StoryID,
		Title:       storyHeader.Title,
		ImageUrl:    storyHeader.ImageURL,
		Description: storyHeader.Description,
	}
}
