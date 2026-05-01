package converter

import (
	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/model"
)

func SavingGoalToResponse(goal *entity.SavingGoal) *model.SavingGoalResponse {
	return &model.SavingGoalResponse{
		ID:          goal.ID,
		MarketID:    goal.MarketID,
		GoalName:    goal.GoalName,
		TargetCoin:  goal.TargetCoin,
		CurrentCoin: goal.CurrentCoin,
	}
}
