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
