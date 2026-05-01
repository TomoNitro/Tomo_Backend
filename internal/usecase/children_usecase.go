package usecase

import (
	"context"

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
	JWT                *helper.JWTHelper
	Redis              *redis.Client
}

func NewChildrenUseCase(db *gorm.DB, log *zap.Logger, validate *validator.Validate, childrenRepository *repository.ChildrenRepository, jwt *helper.JWTHelper, redis *redis.Client) *ChildrenUseCase {
	return &ChildrenUseCase{
		DB:                 db,
		Log:                log,
		Validate:           validate,
		ChildrenRepository: childrenRepository,
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
