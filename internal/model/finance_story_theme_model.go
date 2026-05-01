package model

type StoryFinaceThemeResponse struct {
	Finance []Finance `json:"finance"`
	Story   []Story   `json:"story"`
}
type Finance struct {
	Topic string `json:"topic"`
}
type Story struct {
	Topic     string `json:"topic"`
	FullStory string `json:"fullStory"`
}
