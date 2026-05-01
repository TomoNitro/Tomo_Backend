package model

type StoryHeaderResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ImageUrl    string `json:"image_url"`
	description string `json:"description"`
}
