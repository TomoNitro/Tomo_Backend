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

type DashboardSummaryChild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DashboardSummaryStorySummary struct {
	TotalCompleted int64   `json:"total_completed"`
	SuccessRate    float64 `json:"success_rate"`
}

type DashboardSummarySavingProgress struct {
	GoalID             string  `json:"goal_id"`
	GoalName           string  `json:"goal_name"`
	CurrentCoin        int     `json:"current_coin"`
	TargetCoin         int     `json:"target_coin"`
	ProgressPercentage float64 `json:"progress_percentage"`
}

type GenerateDashboardSummaryWebhookRequest struct {
	Child           DashboardSummaryChild            `json:"child"`
	DecisionSummary DashboardDecisionSummary         `json:"decision_summary"`
	StorySummary    DashboardSummaryStorySummary     `json:"story_summary"`
	SavingProgress  []DashboardSummarySavingProgress `json:"saving_progress"`
	FinancialTrend  []DashboardFinancialTrendItem    `json:"financial_trend"`
}

type DashboardSummaryResponse struct {
	ID               string `json:"id"`
	ChildID          string `json:"child_id"`
	Summary          string `json:"summary"`
	PerformanceLevel string `json:"performance_level"`
	Suggestion       string `json:"suggestion"`
	CreatedAt        string `json:"created_at"`
}
