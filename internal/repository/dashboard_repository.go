package repository

import (
	"time"

	"example.com/tomo/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DashboardRepository struct {
	Log *zap.Logger
}

type DailyCoinTotal struct {
	Date   time.Time `gorm:"column:date"`
	Amount int       `gorm:"column:amount"`
}

func NewDashboardRepository(log *zap.Logger) *DashboardRepository {
	return &DashboardRepository{Log: log}
}

func (r *DashboardRepository) CountDecisionsByWise(db *gorm.DB, childID string) (int64, int64, error) {
	var rows []struct {
		IsWise bool  `gorm:"column:is_wise"`
		Total  int64 `gorm:"column:total"`
	}

	err := db.Model(&entity.Decision{}).
		Select("is_wise, COUNT(*) as total").
		Where("child_id = ?", childID).
		Group("is_wise").
		Scan(&rows).Error
	if err != nil {
		return 0, 0, err
	}

	var wiseCount int64
	var impulsiveCount int64
	for _, row := range rows {
		if row.IsWise {
			wiseCount = row.Total
		} else {
			impulsiveCount = row.Total
		}
	}

	return wiseCount, impulsiveCount, nil
}

func (r *DashboardRepository) CountCompletedStories(db *gorm.DB, childID string) (int64, error) {
	var count int64
	if err := db.Model(&entity.LearningSession{}).
		Where("child_id = ? AND completed_at IS NOT NULL", childID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *DashboardRepository) CountGoodEndingStories(db *gorm.DB, childID, performance string) (int64, error) {
	var count int64
	if err := db.Table("learning_sessions").
		Joins("JOIN story_summary ON learning_sessions.summary_id = story_summary.id").
		Where("learning_sessions.child_id = ? AND learning_sessions.completed_at IS NOT NULL AND story_summary.performance = ?", childID, performance).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *DashboardRepository) FindSavingGoalByChildID(db *gorm.DB, childID string) (*entity.SavingGoal, error) {
	goal := new(entity.SavingGoal)
	if err := db.Where("child_id = ?", childID).First(goal).Error; err != nil {
		return nil, err
	}

	return goal, nil
}

func (r *DashboardRepository) ListDailyCoinTotals(db *gorm.DB, childID string) ([]DailyCoinTotal, error) {
	var totals []DailyCoinTotal
	if err := db.Table("coin_transactions").
		Select("DATE(created_at) as date, COALESCE(SUM(amount), 0) as amount").
		Where("child_id = ?", childID).
		Group("DATE(created_at)").
		Order("DATE(created_at)").
		Scan(&totals).Error; err != nil {
		return nil, err
	}

	return totals, nil
}

func (r *DashboardRepository) CountActiveDays(db *gorm.DB, childID string) (int64, error) {
	var count int64
	if err := db.Model(&entity.Decision{}).
		Where("child_id = ?", childID).
		Distinct("DATE(created_at)").
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *DashboardRepository) FindLatestDashboardSummaryByChildID(db *gorm.DB, childID string) (*entity.DashboardSummary, error) {
	summary := new(entity.DashboardSummary)
	if err := db.Where("child_id = ?", childID).
		Order("created_at DESC").
		First(summary).Error; err != nil {
		return nil, err
	}

	return summary, nil
}
