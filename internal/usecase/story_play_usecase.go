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
	"example.com/tomo/internal/model/converter"
	"example.com/tomo/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	storyChoiceWise      = "wise"
	storyChoiceImpulsive = "impulsive"
	sessionResultGood    = "good"
	sessionResultBad     = "bad"
)

type StoryPlayUseCase struct {
	DB                 *gorm.DB
	Log                *zap.Logger
	Validate           *validator.Validate
	StoryPlayRepo      *repository.StoryPlayRepository
	ChildrenRepository *repository.ChildrenRepository
	SummaryWebhookURL  string
	HTTPClient         *http.Client
}

func NewStoryPlayUseCase(db *gorm.DB, log *zap.Logger, validate *validator.Validate, storyPlayRepo *repository.StoryPlayRepository, childrenRepository *repository.ChildrenRepository, summaryWebhookURL string) *StoryPlayUseCase {
	return &StoryPlayUseCase{
		DB:                 db,
		Log:                log,
		Validate:           validate,
		StoryPlayRepo:      storyPlayRepo,
		ChildrenRepository: childrenRepository,
		SummaryWebhookURL:  summaryWebhookURL,
		HTTPClient:         &http.Client{Timeout: 5 * time.Minute},
	}
}

func (u *StoryPlayUseCase) StartStory(ctx context.Context, actorChildID, storyID string) (*model.StoryPlayStartResponse, error) {
	if actorChildID == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if storyID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "story id is required")
	}

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	child := new(entity.Children)
	if err := u.ChildrenRepository.FindByID(tx, child, actorChildID); err != nil {
		u.Log.Error("failed to find child", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "child not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	storyHeader := new(entity.StoryHeader)
	if err := u.StoryPlayRepo.FindStoryHeaderByID(tx, storyHeader, storyID); err != nil {
		u.Log.Error("failed to find story", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "story not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if storyHeader.RootStoryID == nil || *storyHeader.RootStoryID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "story root node is missing")
	}

	rootNode := new(entity.StoryNode)
	if err := u.StoryPlayRepo.FindStoryNodeByID(tx, rootNode, *storyHeader.RootStoryID); err != nil {
		u.Log.Error("failed to find root story node", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "story root node not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	session := &entity.LearningSession{
		SessionID: uuid.NewString(),
		ChildID:   actorChildID,
		StoryID:   storyHeader.StoryID,
		StartedAt: time.Now(),
	}
	if err := u.StoryPlayRepo.CreateLearningSession(tx, session); err != nil {
		u.Log.Error("failed to create learning session", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit start story transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return &model.StoryPlayStartResponse{
		SessionID: session.SessionID,
		Node:      converter.StoryNodeToPlayResponse(rootNode),
		Progress:  model.StoryPlayProgressResponse{StepsTaken: 0},
	}, nil
}

func (u *StoryPlayUseCase) MakeDecision(ctx context.Context, actorChildID, sessionID string, req *model.StoryDecisionRequest) (*model.StoryPlayDecisionResponse, error) {
	if actorChildID == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if sessionID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "session id is required")
	}
	if err := u.Validate.Struct(req); err != nil {
		u.Log.Error("story decision request is invalid", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	session := new(entity.LearningSession)
	if err := u.StoryPlayRepo.FindLearningSessionByIDAndChildID(tx, session, sessionID, actorChildID); err != nil {
		u.Log.Error("failed to find learning session", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "session not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if session.CompletedAt != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "session already completed")
	}

	currentNode := new(entity.StoryNode)
	if err := u.StoryPlayRepo.FindStoryNodeByID(tx, currentNode, req.NodeID); err != nil {
		u.Log.Error("failed to find current story node", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "story node not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	isWise := req.Choice == storyChoiceWise
	decision := &entity.Decision{
		DecisionID: uuid.NewString(),
		SessionID:  session.SessionID,
		ChildID:    actorChildID,
		NodeID:     currentNode.NodeID,
		IsWise:     isWise,
	}
	if err := u.StoryPlayRepo.CreateDecision(tx, decision); err != nil {
		u.Log.Error("failed to create decision", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	nextNodeID := currentNode.ImpulsiveChoiceNode
	if isWise {
		nextNodeID = currentNode.WiseChoiceNode
	}
	if nextNodeID == nil || *nextNodeID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "next story node is missing")
	}

	nextNode := new(entity.StoryNode)
	if err := u.StoryPlayRepo.FindStoryNodeByID(tx, nextNode, *nextNodeID); err != nil {
		u.Log.Error("failed to find next story node", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "next story node not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	stepsTaken, err := u.StoryPlayRepo.CountDecisionsBySessionID(tx, session.SessionID)
	if err != nil {
		u.Log.Error("failed to count session decisions", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	response := &model.StoryPlayDecisionResponse{
		IsEnd:    nextNode.EndingNode,
		Progress: model.StoryPlayProgressResponse{StepsTaken: stepsTaken},
	}

	if nextNode.EndingNode {
		now := time.Now()
		result := sessionResultBad
		if isWise {
			result = sessionResultGood
		}
		session.CompletedAt = &now
		session.SessionResult = &result

		if err := u.StoryPlayRepo.CompleteLearningSession(tx, session, result); err != nil {
			u.Log.Error("failed to complete learning session", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		storyHeader := new(entity.StoryHeader)
		if err := u.StoryPlayRepo.FindStoryHeaderByID(tx, storyHeader, session.StoryID); err != nil {
			u.Log.Error("failed to find story for summary", zap.Error(err))
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, echo.NewHTTPError(http.StatusNotFound, "story not found")
			}
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		summary := converter.StoryHeaderToPlaySummaryResponse(storyHeader, isWise)
		response.Summary = &summary
	} else {
		node := converter.StoryNodeToPlayResponse(nextNode)
		response.Node = &node
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit decision transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return response, nil
}

func (u *StoryPlayUseCase) GenerateStorySummary(ctx context.Context, actorChildID, sessionID string) (model.GenerateStorySummaryWebhookResponse, error) {
	if actorChildID == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if sessionID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "session id is required")
	}

	session := new(entity.LearningSession)
	if err := u.StoryPlayRepo.FindLearningSessionByIDAndChildID(u.DB.WithContext(ctx), session, sessionID, actorChildID); err != nil {
		u.Log.Error("failed to find learning session for summary generation", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "session not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if session.CompletedAt == nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "session is not completed")
	}

	webhookRequest := model.GenerateStorySummaryWebhookRequest{
		SessionID: session.SessionID,
	}
	requestBody, err := json.Marshal(webhookRequest)
	if err != nil {
		u.Log.Error("failed to marshal generate story summary request", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, u.SummaryWebhookURL, bytes.NewBuffer(requestBody))
	if err != nil {
		u.Log.Error("failed to create generate story summary webhook request", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := u.HTTPClient.Do(httpRequest)
	if err != nil {
		u.Log.Error("failed to call generate story summary webhook", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		u.Log.Error("failed to read generate story summary webhook response", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		u.Log.Error("generate story summary webhook returned non-success status",
			zap.Int("status_code", httpResponse.StatusCode),
			zap.ByteString("response_body", responseBody),
		)
		return nil, echo.NewHTTPError(http.StatusBadGateway, string(responseBody))
	}

	response := make(model.GenerateStorySummaryWebhookResponse)
	if len(responseBody) == 0 {
		return response, nil
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		u.Log.Error("failed to unmarshal generate story summary webhook response", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	return response, nil
}
