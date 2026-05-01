package usecase

import (
	"context"
	"errors"

	"net/http"
	"time"

	"example.com/tomo/internal/entity"
	"example.com/tomo/internal/helper"
	"example.com/tomo/internal/model"
	"example.com/tomo/internal/model/converter"
	"example.com/tomo/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ChildrenUseCase struct {
	DB                 *gorm.DB
	Log                *zap.Logger
	Validate           *validator.Validate
	ChildrenRepository *repository.ChildrenRepository
	CoinRepository     *repository.CoinRepository
	SavingGoalRepo     *repository.SavingGoalRepository
	MarketRepository   *repository.MarketRepository
	JWT                *helper.JWTHelper
	Redis              *redis.Client
}

func NewChildrenUseCase(db *gorm.DB, log *zap.Logger, validate *validator.Validate, childrenRepository *repository.ChildrenRepository, coinRepository *repository.CoinRepository, savingGoalRepo *repository.SavingGoalRepository, marketRepository *repository.MarketRepository, jwt *helper.JWTHelper, redis *redis.Client) *ChildrenUseCase {
	return &ChildrenUseCase{
		DB:                 db,
		Log:                log,
		Validate:           validate,
		ChildrenRepository: childrenRepository,
		CoinRepository:     coinRepository,
		SavingGoalRepo:     savingGoalRepo,
		MarketRepository:   marketRepository,
		JWT:                jwt,
		Redis:              redis,
	}
}
func (u *ChildrenUseCase) ChildrenRegister(ctx context.Context, parentId string, req *model.ChildrenRequest) (resp *model.ChildrenRegisterResponse, err error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.Error("struct request is invalid", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	hashPin, err := helper.HashPassword(req.Pin)
	if err != nil {
		u.Log.Error("Failed to hash pin", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	child := &entity.Children{
		ID:       uuid.NewString(),
		ParentId: parentId,
		Name:     req.Name,
		Pin:      hashPin,
	}
	if err := u.ChildrenRepository.Create(tx, child); err != nil {
		u.Log.Error(err.Error())
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	coin := &entity.CoinTransaction{
		ID:      uuid.NewString(),
		ChildID: child.ID,
		Amount:  75,
	}
	if err := u.CoinRepository.Create(tx, coin); err != nil {
		u.Log.Error("Failed to create initial coin", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	accessToken, err := u.JWT.GenerateToken(helper.BuildAccessTokenClaims(child.ID, helper.ActorTypeChild, child.ParentId))
	if err != nil {
		u.Log.Error("Failed to create token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	refreshToken := uuid.New().String()
	refreshPayload, err := helper.EncodeRefreshTokenPayload(helper.RefreshTokenPayload{
		ActorID:   child.ID,
		ActorType: helper.ActorTypeChild,
		ParentID:  child.ParentId,
	})
	if err != nil {
		u.Log.Error("Failed to encode refresh token payload", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = u.Redis.Set(ctx, refreshToken, refreshPayload, 24*time.Hour).Err()
	if err != nil {
		u.Log.Error("Failed to set refresh token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error(err.Error())
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return converter.ChildrenRegisterToResponse(child, accessToken, refreshToken), nil
}

func (u *ChildrenUseCase) GetChildrenByParent(ctx context.Context, parentID string) (resp []model.ChildrenListResponse, err error) {
	children, err := u.ChildrenRepository.FindByParentID(u.DB.WithContext(ctx), parentID)
	if err != nil {
		u.Log.Error("Failed to get children by parent id", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return converter.ChildrenListToResponse(children), nil
}

func (u *ChildrenUseCase) UpdateChildName(ctx context.Context, childID string, req *model.ChildrenUpdateNameRequest) (resp *model.ChildrenUpdateNameResponse, err error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.Error("struct request is invalid", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	child := new(entity.Children)
	if err := u.ChildrenRepository.FindByID(tx, child, childID); err != nil {
		u.Log.Error("Failed to find child", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusNotFound, "child not found")
	}

	child.Name = req.Name
	if err := u.ChildrenRepository.Update(tx, child); err != nil {
		u.Log.Error("Failed to update child", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("Failed to commit transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return converter.ChildrenUpdateNameToResponse(child), nil
}

func (u *ChildrenUseCase) DeleteChildrenByParent(ctx context.Context, parentID, childID string) (resp *model.ChildrenDeleteResponse, err error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	child := new(entity.Children)
	if err := u.ChildrenRepository.FindByIDAndParentID(tx, child, childID, parentID); err != nil {
		u.Log.Error("Failed to find child by parent id", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusNotFound, "child not found")
	}

	if err := u.ChildrenRepository.Delete(tx, child); err != nil {
		u.Log.Error("Failed to delete child", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("Failed to commit child delete transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return converter.ChildrenDeleteToResponse(child), nil
}

func (u *ChildrenUseCase) ChildrenLogin(ctx context.Context, req *model.ChildrenLoginRequest) (resp *model.ChildrenLoginResponse, err error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.Error("Struct request is invalid", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	child := &entity.Children{}
	if err := u.ChildrenRepository.FindByID(tx, child, req.ChildID); err != nil {
		u.Log.Error("Invalid child pin", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := helper.CompareHashPasswordAndPassword(child.Pin, req.Pin); err != nil {
		u.Log.Error("Invalid credentials", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	accessToken, err := u.JWT.GenerateToken(helper.BuildAccessTokenClaims(child.ID, helper.ActorTypeChild, child.ParentId))
	if err != nil {
		u.Log.Error("Failed to create token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	refreshToken := uuid.New().String()
	refreshPayload, err := helper.EncodeRefreshTokenPayload(helper.RefreshTokenPayload{
		ActorID:   child.ID,
		ActorType: helper.ActorTypeChild,
		ParentID:  child.ParentId,
	})
	if err != nil {
		u.Log.Error("Failed to encode refresh token payload", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = u.Redis.Set(ctx, refreshToken, refreshPayload, 24*time.Hour).Err()
	if err != nil {
		u.Log.Error("Failed to set refresh token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error(err.Error())
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return converter.ChildrenLoginToResponse(child, accessToken, refreshToken), nil
}

func (u *ChildrenUseCase) SetSavingGoal(ctx context.Context, childID, marketID string) (resp *model.SavingGoalResponse, err error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()
	if marketID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "market id is required")
	}

	market := new(entity.Market)
	if err := u.MarketRepository.FindByCondition(tx, market, "id = ?", marketID); err != nil {
		u.Log.Error("Failed to find market", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusNotFound, "market not found")
	}

	coin := new(entity.CoinTransaction)
	if err := u.CoinRepository.FindByChildID(tx, coin, childID); err != nil {
		u.Log.Error("Failed to get coin by child id", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusNotFound, "coin not found")
	}

	goal := new(entity.SavingGoal)
	findErr := u.SavingGoalRepo.FindByChildID(tx, goal, childID)
	if findErr != nil {
		if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			u.Log.Error("Failed to get saving goal", zap.Error(findErr))
			return nil, echo.NewHTTPError(http.StatusBadRequest, findErr.Error())
		}

		goal = &entity.SavingGoal{
			ID:          uuid.NewString(),
			ChildID:     childID,
			MarketID:    market.ID,
			GoalName:    market.Title,
			TargetCoin:  market.Price,
			CurrentCoin: coin.Amount,
		}
		if err := u.SavingGoalRepo.Create(tx, goal); err != nil {
			u.Log.Error("Failed to create saving goal", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		goal.MarketID = market.ID
		goal.GoalName = market.Title
		goal.TargetCoin = market.Price
		goal.CurrentCoin = coin.Amount
		if err := u.SavingGoalRepo.Update(tx, goal); err != nil {
			u.Log.Error("Failed to update saving goal", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("Failed to commit transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return converter.SavingGoalToResponse(goal), nil
}

func (u *ChildrenUseCase) GetChildrenCoin(ctx context.Context, childID string) (resp *model.ChildrenCoinResponse, err error) {
	coin := new(entity.CoinTransaction)
	if err := u.CoinRepository.FindByChildID(u.DB.WithContext(ctx), coin, childID); err != nil {
		u.Log.Error("Failed to get coin by child id", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusNotFound, "coin not found")
	}

	return converter.ChildrenCoinToResponse(coin), nil
}
