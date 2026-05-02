package model

type DashboardDecisionSummary struct {
	WiseCount      int64   `json:"wise_count"`
	ImpulsiveCount int64   `json:"impulsive_count"`
	WisePercentage float64 `json:"wise_percentage"`
}

type DashboardSavingGoal struct {
	GoalName           string  `json:"goal_name"`
	CurrentCoin        int     `json:"current_coin"`
	TargetCoin         int     `json:"target_coin"`
	ProgressPercentage float64 `json:"progress_percentage"`
}

type DashboardStorySummary struct {
	TotalCompletedStories int64   `json:"total_completed_stories"`
	GoodEnding            int64   `json:"good_ending"`
	SuccessRate           float64 `json:"success_rate"`
}

type DashboardFinancialTrendItem struct {
	Date    string `json:"date"`
	Balance int    `json:"balance"`
}

type ParentChildDashboardResponse struct {
	DecisionSummary DashboardDecisionSummary      `json:"decision_summary"`
	SavingGoal      *DashboardSavingGoal          `json:"saving_goal"`
	StorySummary    DashboardStorySummary         `json:"story_summary"`
	FinancialTrend  []DashboardFinancialTrendItem `json:"financial_trend"`
	DaysActive      int64                         `json:"days_active"`
}
