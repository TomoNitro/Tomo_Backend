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

type UserUseCase struct {
	DB             *gorm.DB
	Log            *zap.Logger
	Validate       *validator.Validate
	UserRepository *repository.UserRepository
	JWT            *helper.JWTHelper
	Redis          *redis.Client
}

func NewUserUsecase(db *gorm.DB, log *zap.Logger, validate *validator.Validate, userRepository *repository.UserRepository, jwt *helper.JWTHelper, redis *redis.Client) *UserUseCase {
	return &UserUseCase{
		DB:             db,
		Log:            log,
		Validate:       validate,
		UserRepository: userRepository,
		JWT:            jwt,
		Redis:          redis,
	}
}
func (u *UserUseCase) UserRegister(ctx context.Context, req *model.UserRequest) (resp *model.UserRegisterResponse, err error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.Error("struct request is invalid", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	hashPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		u.Log.Error("Failed to hash password", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	user := &entity.User{
		ID:       uuid.NewString(),
		Email:    req.Email,
		Username: req.Username,
		Password: hashPassword,
	}
	if err := u.UserRepository.Create(tx, user); err != nil {
		u.Log.Error(err.Error())
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	accessToken, err := u.JWT.JWTGenerator(user.ID)
	if err != nil {
		u.Log.Error("Failed to create token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	refreshToken := uuid.New().String()

	err = u.Redis.Set(ctx, refreshToken, user.ID, 24*time.Hour).Err()
	if err != nil {
		u.Log.Error("Failed to set refresh token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error(err.Error())
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return converter.UserRegisterToResponse(user, accessToken, refreshToken), nil
}

func (u *UserUseCase) UserLogin(ctx context.Context, req *model.UserLoginRequest) (resp *model.UserLoginResponse, err error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.Error("Struct request is invalid", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	user := &entity.User{}
	if err := u.UserRepository.Login(tx, user, req.Email); err != nil {
		u.Log.Error("Invalid user email", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := helper.CompareHashPasswordAndPassword(user.Password, req.Password); err != nil {
		u.Log.Error("Invalid credentials", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	accessToken, err := u.JWT.JWTGenerator(user.ID)
	if err != nil {
		u.Log.Error("Failed to create token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	refreshToken := uuid.New().String()

	err = u.Redis.Set(ctx, refreshToken, user.ID, 24*time.Hour).Err()
	if err != nil {
		u.Log.Error("Failed to set refresh token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error(err.Error())
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return converter.UserLoginToResponse(user, accessToken, refreshToken), nil
}
func (u *UserUseCase) RefreshToken(ctx context.Context, req *model.RequestRefreshToken) (resp *model.ResponseRefreshToken, err error) {
	userId, err := u.Redis.Get(ctx, req.RefreshToken).Result()
	if err != nil {
		u.Log.Error("Failed to refresh token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	newAccessToken, err := u.JWT.JWTGenerator(userId)
	if err != nil {
		u.Log.Error("Failed to create new access token", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return converter.RefreshTokenToReponse(newAccessToken), nil
}
