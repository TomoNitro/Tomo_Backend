package usecase

import (
	"context"
	"errors"
	"net/http"
	"time"

	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/model"
	"example.com/tomo/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const expPerLevel = 50

type ProgressUseCase struct {
	DB                *gorm.DB
	Log               *zap.Logger
	ChildProgressRepo *repository.ChildProgressRepository
	BadgeRepo         *repository.BadgeRepository
	ChildBadgeRepo    *repository.ChildBadgeRepository
}

func NewProgressUseCase(db *gorm.DB, log *zap.Logger, childProgressRepo *repository.ChildProgressRepository, badgeRepo *repository.BadgeRepository, childBadgeRepo *repository.ChildBadgeRepository) *ProgressUseCase {
	return &ProgressUseCase{
		DB:                db,
		Log:               log,
		ChildProgressRepo: childProgressRepo,
		BadgeRepo:         badgeRepo,
		ChildBadgeRepo:    childBadgeRepo,
	}
}

func (u *ProgressUseCase) GetChildProgress(ctx context.Context, childID string) (*model.ChildProgressResponse, error) {
	if childID == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	progress := new(entity.ChildProgress)
	totalExp := 0
	level := 1
	if err := u.ChildProgressRepo.FindByChildID(tx, progress, childID); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			u.Log.Error("failed to find child progress", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		totalExp = progress.TotalExp
		level = calculateProgressLevel(totalExp)
		if progress.Level != level {
			progress.Level = level
			if err := u.ChildProgressRepo.Update(tx, progress); err != nil {
				u.Log.Error("failed to update child progress level", zap.Error(err))
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
		}
	}

	badges, err := u.BadgeRepo.FindAll(tx)
	if err != nil {
		u.Log.Error("failed to load badges", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	childBadges, err := u.ChildBadgeRepo.FindByChildID(tx, childID)
	if err != nil {
		u.Log.Error("failed to load child badges", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	childBadgeMap := make(map[string]entity.ChildBadge, len(childBadges))
	for _, badge := range childBadges {
		childBadgeMap[badge.BadgeID] = badge
	}

	responseBadges := make([]model.BadgeResponse, 0, len(badges))
	for _, badge := range badges {
		childBadge, earned := childBadgeMap[badge.ID]
		if !earned && level >= badge.LevelRequired {
			childBadge = entity.ChildBadge{
				ID:       uuid.NewString(),
				ChildID:  childID,
				BadgeID:  badge.ID,
				EarnedAt: time.Now(),
			}
			if err := u.ChildBadgeRepo.Create(tx, &childBadge); err != nil {
				u.Log.Error("failed to create child badge", zap.Error(err))
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			earned = true
			childBadgeMap[badge.ID] = childBadge
		}

		var earnedAt *time.Time
		if earned {
			earnedAt = &childBadge.EarnedAt
		}

		responseBadges = append(responseBadges, model.BadgeResponse{
			ID:            badge.ID,
			Name:          badge.Name,
			Description:   badge.Description,
			LevelRequired: badge.LevelRequired,
			ImageURL:      badge.ImageURL,
			Earned:        earned,
			EarnedAt:      earnedAt,
		})
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit child progress transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	nextLevelExp := level * expPerLevel
	expToNextLevel := nextLevelExp - totalExp
	if expToNextLevel < 0 {
		expToNextLevel = 0
	}

	return &model.ChildProgressResponse{
		TotalExp:       totalExp,
		Level:          level,
		NextLevelExp:   nextLevelExp,
		ExpToNextLevel: expToNextLevel,
		Badges:         responseBadges,
	}, nil
}

func calculateProgressLevel(totalExp int) int {
	if totalExp < 0 {
		return 1
	}

	return (totalExp / expPerLevel) + 1
}
