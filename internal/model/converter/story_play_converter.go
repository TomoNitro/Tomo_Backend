package converter

import (
	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/model"
)

func StoryNodeToPlayResponse(node *entity.StoryNode) model.StoryPlayNodeResponse {
	response := model.StoryPlayNodeResponse{
		NodeID:    node.NodeID,
		AudioText: node.AudioText,
		ImageURL:  node.ImageURL,
		IsEnd:     node.EndingNode,
	}

	if !node.EndingNode {
		response.Choices = &model.StoryPlayChoicesResponse{
			Wise:      node.WiseChoiceText,
			Impulsive: node.ImpulsiveChoiceText,
		}
	}

	return response
}

func StoryHeaderToPlaySummaryResponse(storyHeader *entity.StoryHeader, isWise bool) model.StoryPlaySummaryResponse {
	return model.StoryPlaySummaryResponse{
		Title:       storyHeader.Title,
		Description: storyHeader.Description,
		ImageURL:    storyHeader.ImageURL,
		IsWise:      isWise,
	}
}
