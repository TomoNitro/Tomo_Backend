package middleware

import (
	"net/http"
	"strings"

	"example.com/tomo/internal/helper"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

func AuthMiddleware(jwt *helper.JWTHelper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization format")
			}

			tokenString := parts[1]

			claims, err := jwt.ValidateToken(tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}
			userIDstr, ok := claims["sub"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}
			userId, err := uuid.Parse(userIDstr)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID in token")
			}
			c.Set("user_id", userId.String())
			return next(c)
		}
	}
}
