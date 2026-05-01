package helper

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JWTHelper struct {
	Secret string
	Log    *zap.Logger
}

func NewJWTHelper(secret string, log *zap.Logger) *JWTHelper {
	return &JWTHelper{
		Secret: secret,
		Log:    log,
	}
}

func (j *JWTHelper) JWTGenerator(userId string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
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
