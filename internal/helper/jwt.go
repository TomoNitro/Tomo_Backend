package helper

import (
	"encoding/json"
	"errors"
	"fmt"
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
	claims := jwt.MapClaims{
		"sub":          tokenClaims.Subject,
		"actor_type":   tokenClaims.ActorType,
		"parent_id":    tokenClaims.ParentID,
		"session_type": tokenClaims.SessionType,
		"exp":          time.Now().Add(6 * time.Hour).Unix(),
		"iat":          time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		j.Log.Error(err.Error())
		return tokenString, err
	}
	return tokenString, err
}

func (j *JWTHelper) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.Secret), nil

	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			j.Log.Error("token has expired")
		}
		j.Log.Error("Invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
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
