package helper

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const (
	ActorTypeParent   = "parent"
	ActorTypeChild    = "child"
	SessionTypeAccess = "access"
)

type JWTHelper struct {
	Secret string
	Log    *zap.Logger
}

type TokenClaims struct {
	Subject     string
	ActorType   string
	ParentID    string
	SessionType string
}
type CustomClaims struct {
	ActorType   string `json:"actor_type"`
	ParentID    string `json:"parent_id"`
	SessionType string `json:"session_type"`
	jwt.RegisteredClaims
}
type RefreshTokenPayload struct {
	ActorID   string `json:"actor_id"`
	ActorType string `json:"actor_type"`
	ParentID  string `json:"parent_id"`
}

func NewJWTHelper(secret string, log *zap.Logger) *JWTHelper {
	return &JWTHelper{
		Secret: secret,
		Log:    log,
	}
}

func (j *JWTHelper) GenerateToken(tokenClaims TokenClaims) (string, error) {
	claims := CustomClaims{
		ActorType:   tokenClaims.ActorType,
		ParentID:    tokenClaims.ParentID,
		SessionType: tokenClaims.SessionType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   tokenClaims.Subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(6 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Secret))
}
func (j *JWTHelper) ValidateToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(j.Secret), nil
		},
	)

	if err != nil {
		j.Log.Error("JWT parse error", zap.Error(err))
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
func BuildAccessTokenClaims(actorID, actorType, parentID string) TokenClaims {
	if parentID == "" {
		parentID = actorID
	}

	return TokenClaims{
		Subject:     actorID,
		ActorType:   actorType,
		ParentID:    parentID,
		SessionType: SessionTypeAccess,
	}
}

func ParseTokenClaims(claims jwt.MapClaims) (*TokenClaims, error) {
	subject, ok := claims["sub"].(string)
	if !ok || subject == "" {
		return nil, errors.New("invalid token subject")
	}
	actorType, ok := claims["actor_type"].(string)
	if !ok || actorType == "" {
		return nil, errors.New("invalid actor type")
	}
	parentID, ok := claims["parent_id"].(string)
	if !ok || parentID == "" {
		return nil, errors.New("invalid parent id")
	}
	sessionType, ok := claims["session_type"].(string)
	if !ok || sessionType == "" {
		return nil, errors.New("invalid session type")
	}

	return &TokenClaims{
		Subject:     subject,
		ActorType:   actorType,
		ParentID:    parentID,
		SessionType: sessionType,
	}, nil
}

func EncodeRefreshTokenPayload(payload RefreshTokenPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func DecodeRefreshTokenPayload(value string) (*RefreshTokenPayload, error) {
	payload := new(RefreshTokenPayload)
	if err := json.Unmarshal([]byte(value), payload); err != nil {
		return nil, err
	}
	if payload.ActorID == "" || payload.ActorType == "" || payload.ParentID == "" {
		return nil, errors.New("invalid refresh token payload")
	}

	return payload, nil
}
