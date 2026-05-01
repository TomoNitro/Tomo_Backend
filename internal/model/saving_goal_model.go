package model

type SavingGoalResponse struct {
	ID          string `json:"id"`
	MarketID    string `json:"market_id"`
	GoalName    string `json:"goal_name"`
	TargetCoin  int    `json:"target_coin"`
	CurrentCoin int    `json:"current_coin"`
}
