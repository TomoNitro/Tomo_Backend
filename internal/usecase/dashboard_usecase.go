package usecase

import (
	"context"
	"errors"
	"net/http"

	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/model"
	"example.com/tomo/internal/repository"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const storySummaryPerformanceWise = "wise"

type DashboardUseCase struct {
	DB                  *gorm.DB
	Log                 *zap.Logger
	ChildrenRepository  *repository.ChildrenRepository
	DashboardRepository *repository.DashboardRepository
}

func NewDashboardUseCase(db *gorm.DB, log *zap.Logger, childrenRepository *repository.ChildrenRepository, dashboardRepository *repository.DashboardRepository) *DashboardUseCase {
	return &DashboardUseCase{
		DB:                  db,
		Log:                 log,
		ChildrenRepository:  childrenRepository,
		DashboardRepository: dashboardRepository,
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

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit dashboard transaction", zap.Error(err))
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
