package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	storyChoiceWise            = "wise"
	storyChoiceImpulsive       = "impulsive"
	expSourceStoryComplete     = "story_complete"
	summaryWebhookMaxAttempts  = 3
	summaryWebhookBaseBackoff  = 500 * time.Millisecond
	summaryWebhookLogBodyLimit = 2048
)

type StoryPlayUseCase struct {
	DB                  *gorm.DB
	Log                 *zap.Logger
	Validate            *validator.Validate
	StoryPlayRepo       *repository.StoryPlayRepository
	ChildrenRepository  *repository.ChildrenRepository
	ChildProgressRepo   *repository.ChildProgressRepository
	ExpTransactionRepo  *repository.ExpTransactionRepository
	CoinRepository      *repository.CoinRepository
	SummaryWebhookURL   string
	NodeAudioWebhookURL string
	HTTPClient          *http.Client
}

func NewStoryPlayUseCase(db *gorm.DB, log *zap.Logger, validate *validator.Validate, storyPlayRepo *repository.StoryPlayRepository, childrenRepository *repository.ChildrenRepository, childProgressRepo *repository.ChildProgressRepository, expTransactionRepo *repository.ExpTransactionRepository, coinRepository *repository.CoinRepository, summaryWebhookURL, nodeAudioWebhookURL string) *StoryPlayUseCase {
	return &StoryPlayUseCase{
		DB:                  db,
		Log:                 log,
		Validate:            validate,
		StoryPlayRepo:       storyPlayRepo,
		ChildrenRepository:  childrenRepository,
		ChildProgressRepo:   childProgressRepo,
		ExpTransactionRepo:  expTransactionRepo,
		CoinRepository:      coinRepository,
		SummaryWebhookURL:   summaryWebhookURL,
		NodeAudioWebhookURL: nodeAudioWebhookURL,
		HTTPClient:          &http.Client{Timeout: 5 * time.Minute},
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

	completed, err := u.StoryPlayRepo.HasCompletedStory(tx, actorChildID, storyHeader.StoryID)
	if err != nil {
		u.Log.Error("failed to check completed story", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if completed {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "story already completed")
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
		session.CompletedAt = &now

		if err := u.StoryPlayRepo.CompleteLearningSession(tx, session); err != nil {
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

func (u *StoryPlayUseCase) GenerateStoryNodeAudio(ctx context.Context, actorChildID, nodeID string) (*model.StoryNodeAudioResponse, error) {
	if actorChildID == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if nodeID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "node id is required")
	}

	storyNode := new(entity.StoryNode)
	if err := u.StoryPlayRepo.FindStoryNodeByID(u.DB.WithContext(ctx), storyNode, nodeID); err != nil {
		u.Log.Error("failed to find story node for audio generation", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "story node not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	webhookRequest := model.GenerateStoryNodeAudioWebhookRequest{
		NodeID: storyNode.NodeID,
		Text:   storyNode.AudioText,
	}
	requestBody, err := json.Marshal(webhookRequest)
	if err != nil {
		u.Log.Error("failed to marshal generate story node audio request", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, u.NodeAudioWebhookURL, bytes.NewBuffer(requestBody))
	if err != nil {
		u.Log.Error("failed to create generate story node audio webhook request", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := u.HTTPClient.Do(httpRequest)
	if err != nil {
		u.Log.Error("failed to call generate story node audio webhook", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		u.Log.Error("failed to read generate story node audio webhook response", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		u.Log.Error("generate story node audio webhook returned non-success status",
			zap.Int("status_code", httpResponse.StatusCode),
			zap.ByteString("response_body", responseBody),
		)
		return nil, echo.NewHTTPError(http.StatusBadGateway, string(responseBody))
	}
	if len(responseBody) == 0 {
		return nil, echo.NewHTTPError(http.StatusBadGateway, "generate story node audio webhook returned empty response")
	}

	webhookResponse := make(model.GenerateStoryNodeAudioWebhookResponse)
	if err := json.Unmarshal(responseBody, &webhookResponse); err != nil {
		u.Log.Error("failed to unmarshal generate story node audio webhook response", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}

	audioURL := audioURLFromWebhookResponse(webhookResponse)
	if audioURL == "" {
		u.Log.Error("generate story node audio webhook returned invalid payload")
		return nil, echo.NewHTTPError(http.StatusBadGateway, "invalid story node audio response")
	}

	return &model.StoryNodeAudioResponse{
		NodeID:   storyNode.NodeID,
		AudioURL: audioURL,
	}, nil
}

func (u *StoryPlayUseCase) GenerateStorySummary(ctx context.Context, actorChildID, sessionID string) (*model.StorySummaryRewardResponse, error) {
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
	rewardCount, err := u.ExpTransactionRepo.CountBySourceAndReferenceID(u.DB.WithContext(ctx), expSourceStoryComplete, session.SessionID)
	if err != nil {
		u.Log.Error("failed to check existing story summary reward", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if rewardCount > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "story summary reward already claimed")
	}

	webhookRequest := model.GenerateStorySummaryWebhookRequest{
		SessionID: session.SessionID,
	}
	requestBody, err := json.Marshal(webhookRequest)
	if err != nil {
		u.Log.Error("failed to marshal generate story summary request", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	responseBody, err := u.callSummaryWebhook(ctx, requestBody, session.SessionID, actorChildID)
	if err != nil {
		fallbackResponse, fallbackErr := u.buildSummaryResponseFromDB(ctx, session.SessionID, actorChildID)
		if fallbackErr == nil {
			u.Log.Warn("summary webhook failed; returning summary from database",
				zap.String("session_id", session.SessionID),
				zap.String("child_id", actorChildID),
				zap.Error(err),
			)
			return fallbackResponse, nil
		}
		return nil, err
	}

	webhookResponse := make(model.GenerateStorySummaryWebhookResponse)
	if err := json.Unmarshal(responseBody, &webhookResponse); err != nil {
		u.Log.Error("failed to unmarshal generate story summary webhook response", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, "invalid story summary response")
	}

	exp, err := intFromWebhookValue(webhookResponse["exp"])
	if err != nil {
		u.Log.Error("failed to parse exp reward", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, "invalid exp reward from summary service")
	}
	coins, err := intFromWebhookValue(webhookResponse["coins"])
	if err != nil {
		u.Log.Error("failed to parse coin reward", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadGateway, "invalid coin reward from summary service")
	}

	summaryID := stringFromWebhookValue(webhookResponse["id"])
	summaryTitle := stringFromWebhookValue(webhookResponse["title"])
	summaryDescription := stringFromWebhookValue(webhookResponse["description"])
	summaryPerformance := stringFromWebhookValue(webhookResponse["performance"])
	if summaryID == "" || summaryPerformance == "" {
		u.Log.Error("generate story summary webhook returned invalid payload")
		return nil, echo.NewHTTPError(http.StatusBadGateway, "invalid story summary response from summary service")
	}

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	session = new(entity.LearningSession)
	if err := u.StoryPlayRepo.FindLearningSessionByIDAndChildID(tx, session, sessionID, actorChildID); err != nil {
		u.Log.Error("failed to find learning session before saving reward", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "session not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if session.CompletedAt == nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "session is not completed")
	}
	rewardCount, err = u.ExpTransactionRepo.CountBySourceAndReferenceID(tx, expSourceStoryComplete, session.SessionID)
	if err != nil {
		u.Log.Error("failed to recheck existing story summary reward", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if rewardCount > 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "story summary reward already claimed")
	}

	summary := &entity.StorySummary{
		ID:          summaryID,
		Title:       summaryTitle,
		Description: summaryDescription,
		Performance: summaryPerformance,
	}
	if err := u.StoryPlayRepo.CreateStorySummary(tx, summary); err != nil {
		u.Log.Error("failed to create story summary", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := u.StoryPlayRepo.UpdateLearningSessionSummaryID(tx, session.SessionID, summary.ID); err != nil {
		u.Log.Error("failed to update learning session summary", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	progress := new(entity.ChildProgress)
	progressErr := u.ChildProgressRepo.FindByChildID(tx, progress, actorChildID)
	if progressErr != nil {
		if !errors.Is(progressErr, gorm.ErrRecordNotFound) {
			u.Log.Error("failed to find child progress", zap.Error(progressErr))
			return nil, echo.NewHTTPError(http.StatusBadRequest, progressErr.Error())
		}

		progress = &entity.ChildProgress{
			ChildID:  actorChildID,
			TotalExp: exp,
			Level:    calculateChildLevel(exp),
		}
		if err := u.ChildProgressRepo.Create(tx, progress); err != nil {
			u.Log.Error("failed to create child progress", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		progress.TotalExp += exp
		progress.Level = calculateChildLevel(progress.TotalExp)
		if err := u.ChildProgressRepo.Update(tx, progress); err != nil {
			u.Log.Error("failed to update child progress", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	referenceID := session.SessionID
	expTransaction := &entity.ExpTransaction{
		ID:          uuid.NewString(),
		ChildID:     actorChildID,
		Amount:      exp,
		Source:      expSourceStoryComplete,
		ReferenceID: &referenceID,
	}
	if err := u.ExpTransactionRepo.Create(tx, expTransaction); err != nil {
		u.Log.Error("failed to create exp transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	coinTransaction := &entity.CoinTransaction{
		ID:      uuid.NewString(),
		ChildID: actorChildID,
		Amount:  coins,
	}
	if err := u.CoinRepository.Create(tx, coinTransaction); err != nil {
		u.Log.Error("failed to create coin transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	totalCoins, err := u.CoinRepository.SumAmountByChildID(tx, actorChildID)
	if err != nil {
		u.Log.Error("failed to get child coin total", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit story summary reward transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return &model.StorySummaryRewardResponse{
		ID:          stringFromWebhookValue(webhookResponse["id"]),
		Title:       stringFromWebhookValue(webhookResponse["title"]),
		Description: stringFromWebhookValue(webhookResponse["description"]),
		Performance: stringFromWebhookValue(webhookResponse["performance"]),
		Exp:         exp,
		Coins:       coins,
		TotalExp:    progress.TotalExp,
		Level:       progress.Level,
		TotalCoins:  totalCoins,
	}, nil
}

func (u *StoryPlayUseCase) buildSummaryResponseFromDB(ctx context.Context, sessionID, actorChildID string) (*model.StorySummaryRewardResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	session := new(entity.LearningSession)
	if err := u.StoryPlayRepo.FindLearningSessionByIDAndChildID(tx, session, sessionID, actorChildID); err != nil {
		u.Log.Error("failed to find learning session for summary fallback", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "session not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if session.SummaryID == nil || *session.SummaryID == "" {
		return nil, echo.NewHTTPError(http.StatusBadGateway, "story summary not linked to session")
	}

	summary := new(entity.StorySummary)
	if err := u.StoryPlayRepo.FindStorySummaryByID(tx, summary, *session.SummaryID); err != nil {
		u.Log.Error("failed to find story summary for fallback", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, echo.NewHTTPError(http.StatusNotFound, "story summary not found")
		}
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	expAmount := 0
	expTransaction := new(entity.ExpTransaction)
	if err := u.ExpTransactionRepo.FindBySourceAndReferenceID(tx, expTransaction, expSourceStoryComplete, session.SessionID); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			u.Log.Error("failed to find exp transaction for fallback", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		expAmount = expTransaction.Amount
	}

	progressTotal := 0
	progressLevel := 1
	progress := new(entity.ChildProgress)
	if err := u.ChildProgressRepo.FindByChildID(tx, progress, actorChildID); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			u.Log.Error("failed to find child progress for fallback", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		progressTotal = progress.TotalExp
		progressLevel = progress.Level
	}

	totalCoins, err := u.CoinRepository.SumAmountByChildID(tx, actorChildID)
	if err != nil {
		u.Log.Error("failed to get child coin total for fallback", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.Error("failed to commit summary fallback transaction", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return &model.StorySummaryRewardResponse{
		ID:          summary.ID,
		Title:       summary.Title,
		Description: summary.Description,
		Performance: summary.Performance,
		Exp:         expAmount,
		Coins:       0,
		TotalExp:    progressTotal,
		Level:       progressLevel,
		TotalCoins:  totalCoins,
	}, nil
}

func (u *StoryPlayUseCase) callSummaryWebhook(ctx context.Context, requestBody []byte, sessionID, childID string) ([]byte, error) {
	backoff := summaryWebhookBaseBackoff
	for attempt := 1; attempt <= summaryWebhookMaxAttempts; attempt++ {
		httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, u.SummaryWebhookURL, bytes.NewBuffer(requestBody))
		if err != nil {
			u.Log.Error("failed to create generate story summary webhook request", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		httpRequest.Header.Set("Content-Type", "application/json")

		httpResponse, err := u.HTTPClient.Do(httpRequest)
		if err != nil {
			u.Log.Error("failed to call generate story summary webhook",
				zap.Error(err),
				zap.Int("attempt", attempt),
				zap.String("session_id", sessionID),
				zap.String("child_id", childID),
			)
			if attempt < summaryWebhookMaxAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, echo.NewHTTPError(http.StatusBadGateway, "story summary service unavailable")
		}

		responseBody, readErr := io.ReadAll(httpResponse.Body)
		httpResponse.Body.Close()
		if readErr != nil {
			u.Log.Error("failed to read generate story summary webhook response",
				zap.Error(readErr),
				zap.Int("attempt", attempt),
				zap.String("session_id", sessionID),
				zap.String("child_id", childID),
			)
			if attempt < summaryWebhookMaxAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, echo.NewHTTPError(http.StatusBadGateway, "story summary service response read failed")
		}

		if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
			u.Log.Error("generate story summary webhook returned non-success status",
				zap.Int("status_code", httpResponse.StatusCode),
				zap.ByteString("response_body", truncateLogBytes(responseBody, summaryWebhookLogBodyLimit)),
				zap.Int("attempt", attempt),
				zap.String("session_id", sessionID),
				zap.String("child_id", childID),
			)
			if shouldRetrySummaryStatus(httpResponse.StatusCode) && attempt < summaryWebhookMaxAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("story summary service returned status %d", httpResponse.StatusCode))
		}

		if len(responseBody) == 0 {
			u.Log.Error("generate story summary webhook returned empty response",
				zap.Int("status_code", httpResponse.StatusCode),
				zap.Int("attempt", attempt),
				zap.String("session_id", sessionID),
				zap.String("child_id", childID),
			)
			if attempt < summaryWebhookMaxAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, echo.NewHTTPError(http.StatusBadGateway, "story summary service returned empty response")
		}

		return responseBody, nil
	}

	return nil, echo.NewHTTPError(http.StatusBadGateway, "story summary service unavailable")
}

func intFromWebhookValue(value interface{}) (int, error) {
	switch typedValue := value.(type) {
	case string:
		return strconv.Atoi(typedValue)
	case float64:
		return int(typedValue), nil
	case int:
		return typedValue, nil
	default:
		return 0, errors.New("value is not an integer")
	}
}

func stringFromWebhookValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch typedValue := value.(type) {
	case string:
		return typedValue
	case float64:
		return strconv.FormatFloat(typedValue, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typedValue)
	default:
		return ""
	}
}

func audioURLFromWebhookResponse(response model.GenerateStoryNodeAudioWebhookResponse) string {
	for _, key := range []string{"audio_url", "audioUrl", "url", "audio"} {
		if audioURL := stringFromWebhookValue(response[key]); audioURL != "" {
			return audioURL
		}
	}

	return ""
}

func shouldRetrySummaryStatus(status int) bool {
	if status == http.StatusRequestTimeout || status == http.StatusTooManyRequests {
		return true
	}
	return status >= http.StatusInternalServerError
}

func truncateLogBytes(value []byte, limit int) []byte {
	if len(value) <= limit {
		return value
	}

	trimmed := make([]byte, limit+15)
	copy(trimmed, value[:limit])
	copy(trimmed[limit:], []byte("... (truncated)"))
	return trimmed
}

func calculateChildLevel(totalExp int) int {
	if totalExp < 0 {
		return 1
	}

	return (totalExp / 100) + 1
}
