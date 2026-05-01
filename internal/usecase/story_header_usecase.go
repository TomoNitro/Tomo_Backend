package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"example.com/tomo/internal/model"
	"example.com/tomo/internal/model/converter"
	"example.com/tomo/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type StoryHeaderUseCase struct {
	DB                    *gorm.DB
	Log                   *zap.Logger
	Validate              *validator.Validate
	StoryHeaderRepository *repository.StoryHeaderRepository
	StoryWebhookURL       string
	HTTPClient            *http.Client
}

func NewStoryHeaderUseCase(db *gorm.DB, logger *zap.Logger, validate *validator.Validate, storyHeaderRepository *repository.StoryHeaderRepository, storyWebhookURL string) *StoryHeaderUseCase {
	return &StoryHeaderUseCase{
		DB:                    db,
		Log:                   logger,
		Validate:              validate,
		StoryHeaderRepository: storyHeaderRepository,
		StoryWebhookURL:       storyWebhookURL,
		HTTPClient:            &http.Client{Timeout: 5 * time.Minute}}
}

func (s *StoryHeaderUseCase) GetAllStoryByParentId(ctx context.Context, parentId string) (*[]model.StoryHeaderResponse, error) {
	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	storyHeader, err := s.StoryHeaderRepository.GetAllStoryHeaderByParentId(tx, parentId)
	if err != nil {
		s.Log.Error("Failed to get story header by parent id", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())

	}
	if err := tx.Commit().Error; err != nil {
		s.Log.Error("Failed to commit transaction")
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	response := make([]model.StoryHeaderResponse, len(*storyHeader))
	for i, storyHeaders := range *storyHeader {
		response[i] = *converter.StoryHeaderToResponse(&storyHeaders)
	}
	return &response, nil
}

func (s *StoryHeaderUseCase) CreateStory(ctx context.Context, parentID string, req *model.CreateStoryRequest) (resp model.CreateStoryWebhookResponse, err error) {
	if err := s.Validate.Struct(req); err != nil {
		s.Log.Error("create story request is invalid", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	webhookRequest := model.CreateStoryWebhookRequest{
		ParentID:     parentID,
		Topic:        req.Topic,
		Story:        req.Story,
		CustomPrompt: req.CustomPrompt,
	}

	requestBody, err := json.Marshal(webhookRequest)
	if err != nil {
		s.Log.Error("failed to marshal create story request", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, s.StoryWebhookURL, bytes.NewBuffer(requestBody))
	if err != nil {
		s.Log.Error("failed to create story webhook request", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := s.HTTPClient.Do(httpRequest)
	if err != nil {
		s.Log.Error("failed to call story webhook", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		s.Log.Error("failed to read story webhook response", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		s.Log.Error("story webhook returned non-success status",
			zap.Int("status_code", httpResponse.StatusCode),
			zap.ByteString("response_body", responseBody),
		)
		return nil, echo.NewHTTPError(http.StatusBadGateway, string(responseBody))
	}

	response := make(model.CreateStoryWebhookResponse)
	if len(responseBody) == 0 {
		return response, nil
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		s.Log.Error("failed to unmarshal story webhook response", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	return response, nil
}
