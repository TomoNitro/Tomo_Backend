package entity

type StoryNode struct {
	NodeID              string  `gorm:"column:node_id;primaryKey"`
	AudioText           string  `gorm:"column:audio_text"`
	ImageURL            string  `gorm:"column:image_url"`
	IsRoot              bool    `gorm:"column:is_root"`
	EndingNode          bool    `gorm:"column:ending_node"`
	WiseChoiceNode      *string `gorm:"column:wise_choice_node"`
	ImpulsiveChoiceNode *string `gorm:"column:impulsive_choice_node"`
	WiseChoiceText      string  `gorm:"column:wise_choice_text"`
	ImpulsiveChoiceText string  `gorm:"column:impulsive_choice_text"`
}

func (s *StoryNode) TableName() string {
	return "story_node"
}
