package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/model"
	"example.com/tomo/internal/repository"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const storySummaryPerformanceWise = "wise"

type DashboardUseCase struct {
	DB                         *gorm.DB
	Log                        *zap.Logger
	ChildrenRepository         *repository.ChildrenRepository
	DashboardRepository        *repository.DashboardRepository
	DashboardSummaryWebhookURL string
	HTTPClient                 *http.Client
}

func NewDashboardUseCase(db *gorm.DB, log *zap.Logger, childrenRepository *repository.ChildrenRepository, dashboardRepository *repository.DashboardRepository, dashboardSummaryWebhookURL string) *DashboardUseCase {
	return &DashboardUseCase{
		DB:                         db,
		Log:                        log,
		ChildrenRepository:         childrenRepository,
		DashboardRepository:        dashboardRepository,
		DashboardSummaryWebhookURL: dashboardSummaryWebhookURL,
		HTTPClient:                 &http.Client{Timeout: 5 * time.Minute},
	}
}

func (u *DashboardUseCase) GetChildDashboard(ctx context.Context, parentID, childID string) (*model.ParentChildDashboardResponse, error) {
	if childID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "child id is required")
	}

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	child := new(entity.Children)
	if err := u.ChildrenRepository.FindByIDAndParentID(tx, child, childID, parentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "child not found")
		}
		u.Log.Error("failed to find child by parent", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	response, err := u.buildChildDashboardResponse(tx, childID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit dashboard transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return response, nil
}

func (u *DashboardUseCase) GenerateChildDashboardSummary(ctx context.Context, parentID, childID string) (*model.DashboardSummaryResponse, error) {
	if childID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "child id is required")
	}

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	child := new(entity.Children)
	if err := u.ChildrenRepository.FindByIDAndParentID(tx, child, childID, parentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "child not found")
		}
		u.Log.Error("failed to find child by parent", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	dashboard, err := u.buildChildDashboardResponse(tx, childID)
	if err != nil {
		return nil, err
	}

	savingProgress := make([]model.DashboardSummarySavingProgress, 0, 1)
	if dashboard.SavingGoal != nil {
		goal, err := u.DashboardRepository.FindSavingGoalByChildID(tx, childID)
		if err != nil {
			u.Log.Error("failed to find saving goal for dashboard summary", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		savingProgress = append(savingProgress, model.DashboardSummarySavingProgress{
			GoalID:             goal.ID,
			GoalName:           dashboard.SavingGoal.GoalName,
			CurrentCoin:        dashboard.SavingGoal.CurrentCoin,
			TargetCoin:         dashboard.SavingGoal.TargetCoin,
			ProgressPercentage: dashboard.SavingGoal.ProgressPercentage,
		})
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit dashboard summary transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	webhookRequest := model.GenerateDashboardSummaryWebhookRequest{
		Child: model.DashboardSummaryChild{
			ID:   child.ID,
			Name: child.Name,
		},
		DecisionSummary: dashboard.DecisionSummary,
		StorySummary: model.DashboardSummaryStorySummary{
			TotalCompleted: dashboard.StorySummary.TotalCompletedStories,
			SuccessRate:    dashboard.StorySummary.SuccessRate,
		},
		SavingProgress: savingProgress,
		FinancialTrend: dashboard.FinancialTrend,
	}

	requestBody, err := json.Marshal(webhookRequest)
	if err != nil {
		u.Log.Error("failed to marshal generate dashboard summary request", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, u.DashboardSummaryWebhookURL, bytes.NewBuffer(requestBody))
	if err != nil {
		u.Log.Error("failed to create generate dashboard summary webhook request", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := u.HTTPClient.Do(httpRequest)
	if err != nil {
		u.Log.Error("failed to call generate dashboard summary webhook", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		u.Log.Error("failed to read generate dashboard summary webhook response", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		u.Log.Error("generate dashboard summary webhook returned non-success status",
			zap.Int("status_code", httpResponse.StatusCode),
			zap.ByteString("response_body", responseBody),
		)
		return nil, echo.NewHTTPError(http.StatusBadGateway, string(responseBody))
	}
	if len(responseBody) == 0 {
		return nil, echo.NewHTTPError(http.StatusBadGateway, "generate dashboard summary webhook returned empty response")
	}

	response := new(model.DashboardSummaryResponse)
	if err := json.Unmarshal(responseBody, response); err != nil {
		u.Log.Error("failed to unmarshal generate dashboard summary webhook response", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	if response.ID == "" || response.ChildID == "" || response.Summary == "" {
		u.Log.Error("generate dashboard summary webhook returned invalid payload")
		return nil, echo.NewHTTPError(http.StatusBadGateway, "invalid dashboard summary response")
	}

	return response, nil
}

func (u *DashboardUseCase) GetLatestChildDashboardSummary(ctx context.Context, parentID, childID string) (*model.DashboardSummaryResponse, error) {
	if childID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "child id is required")
	}

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	child := new(entity.Children)
	if err := u.ChildrenRepository.FindByIDAndParentID(tx, child, childID, parentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "child not found")
		}
		u.Log.Error("failed to find child by parent", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	summary, err := u.DashboardRepository.FindLatestDashboardSummaryByChildID(tx, childID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "dashboard summary not found")
		}
		u.Log.Error("failed to find latest dashboard summary", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit get dashboard summary transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return dashboardSummaryToResponse(summary), nil
}

func (u *DashboardUseCase) buildChildDashboardResponse(tx *gorm.DB, childID string) (*model.ParentChildDashboardResponse, error) {
	wiseCount, impulsiveCount, err := u.DashboardRepository.CountDecisionsByWise(tx, childID)
	if err != nil {
		u.Log.Error("failed to count decisions", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	totalDecisions := wiseCount + impulsiveCount
	wisePercentage := 0.0
	if totalDecisions > 0 {
		wisePercentage = (float64(wiseCount) / float64(totalDecisions)) * 100
	}

	var savingGoalResponse *model.DashboardSavingGoal
	goal, err := u.DashboardRepository.FindSavingGoalByChildID(tx, childID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			u.Log.Error("failed to find saving goal", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		progressPercentage := 0.0
		if goal.TargetCoin > 0 {
			progressPercentage = (float64(goal.CurrentCoin) / float64(goal.TargetCoin)) * 100
		}

		savingGoalResponse = &model.DashboardSavingGoal{
			GoalName:           goal.GoalName,
			CurrentCoin:        goal.CurrentCoin,
			TargetCoin:         goal.TargetCoin,
			ProgressPercentage: progressPercentage,
		}
	}

	totalCompleted, err := u.DashboardRepository.CountCompletedStories(tx, childID)
	if err != nil {
		u.Log.Error("failed to count completed stories", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	goodEnding, err := u.DashboardRepository.CountGoodEndingStories(tx, childID, storySummaryPerformanceWise)
	if err != nil {
		u.Log.Error("failed to count good ending stories", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	successRate := 0.0
	if totalCompleted > 0 {
		successRate = (float64(goodEnding) / float64(totalCompleted)) * 100
	}

	dailyTotals, err := u.DashboardRepository.ListDailyCoinTotals(tx, childID)
	if err != nil {
		u.Log.Error("failed to list daily coin totals", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	trend := make([]model.DashboardFinancialTrendItem, 0, len(dailyTotals))
	balance := 0
	for _, daily := range dailyTotals {
		balance += daily.Amount
		trend = append(trend, model.DashboardFinancialTrendItem{
			Date:    daily.Date.Format("2006-01-02"),
			Balance: balance,
		})
	}

	daysActive, err := u.DashboardRepository.CountActiveDays(tx, childID)
	if err != nil {
		u.Log.Error("failed to count active days", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return &model.ParentChildDashboardResponse{
		DecisionSummary: model.DashboardDecisionSummary{
			WiseCount:      wiseCount,
			ImpulsiveCount: impulsiveCount,
			WisePercentage: wisePercentage,
		},
		SavingGoal: savingGoalResponse,
		StorySummary: model.DashboardStorySummary{
			TotalCompletedStories: totalCompleted,
			GoodEnding:            goodEnding,
			SuccessRate:           successRate,
		},
		FinancialTrend: trend,
		DaysActive:     daysActive,
	}, nil
}

func dashboardSummaryToResponse(summary *entity.DashboardSummary) *model.DashboardSummaryResponse {
	return &model.DashboardSummaryResponse{
		ID:               summary.ID,
		ChildID:          summary.ChildID,
		Summary:          summary.Summary,
		PerformanceLevel: summary.PerformanceLevel,
		Suggestion:       summary.Suggestion,
		CreatedAt:        summary.CreatedAt.Format(time.RFC3339Nano),
	}
}
