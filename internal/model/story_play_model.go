package model

type StoryDecisionRequest struct {
	NodeID string `json:"node_id" validate:"required"`
	Choice string `json:"choice" validate:"required,oneof=wise impulsive"`
}

type StoryPlayStartResponse struct {
	SessionID string                    `json:"session_id"`
	Node      StoryPlayNodeResponse     `json:"node"`
	Progress  StoryPlayProgressResponse `json:"progress"`
}

type StoryPlayDecisionResponse struct {
	IsEnd    bool                      `json:"is_end"`
	Node     *StoryPlayNodeResponse    `json:"node,omitempty"`
	Summary  *StoryPlaySummaryResponse `json:"summary,omitempty"`
	Progress StoryPlayProgressResponse `json:"progress"`
}

type StoryPlayNodeResponse struct {
	NodeID    string                    `json:"node_id"`
	AudioText string                    `json:"audio_text"`
	ImageURL  string                    `json:"image_url"`
	IsEnd     bool                      `json:"is_end"`
	Choices   *StoryPlayChoicesResponse `json:"choices,omitempty"`
}

type StoryPlayChoicesResponse struct {
	Wise      string `json:"wise"`
	Impulsive string `json:"impulsive"`
}

type StoryPlaySummaryResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	IsWise      bool   `json:"is_wise"`
}

type StoryPlayProgressResponse struct {
	StepsTaken int64 `json:"steps_taken"`
}

type GenerateStorySummaryWebhookRequest struct {
	SessionID string `json:"session_id"`
}

type GenerateStorySummaryWebhookResponse map[string]interface{}

type GenerateStoryNodeAudioWebhookRequest struct {
	NodeID string `json:"node_id"`
	Text   string `json:"text"`
}

type GenerateStoryNodeAudioWebhookResponse map[string]interface{}

type StoryNodeAudioResponse struct {
	NodeID   string `json:"node_id"`
	AudioURL string `json:"audio_url"`
}

type StorySummaryRewardResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Performance string `json:"performance"`
	Exp         int    `json:"exp"`
	Coins       int    `json:"coins"`
	TotalExp    int    `json:"total_exp"`
	Level       int    `json:"level"`
	TotalCoins  int    `json:"total_coins"`
}
