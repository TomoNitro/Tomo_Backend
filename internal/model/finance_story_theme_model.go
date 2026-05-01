package model

type StoryFinaceThemeResponse struct {
	Finance []Finance `json:"finance"`
	Story   []Story   `json:"story"`
}
type Finance struct {
	Topic string `json:"topic"`
}
type Story struct {
	Title     string `json:"title"`
	FullStory string `json:"fullStory"`
}
