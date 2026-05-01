package model

type StoryHeaderResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ImageUrl    string `json:"image_url"`
	Description string `json:"description"`
}

type CreateStoryRequest struct {
	Topic        string            `json:"topic" validate:"required,max=255"`
	Story        CreateStoryDetail `json:"story" validate:"required"`
	CustomPrompt string            `json:"customPrompt" validate:"required"`
}

type CreateStoryDetail struct {
	Title     string `json:"title" validate:"required,max=255"`
	FullStory string `json:"full_story" validate:"required"`
}

type CreateStoryWebhookRequest struct {
	ParentID     string            `json:"parent_id"`
	Topic        string            `json:"topic"`
	Story        CreateStoryDetail `json:"story"`
	CustomPrompt string            `json:"custom prompt"`
}

type CreateStoryWebhookResponse map[string]interface{}
