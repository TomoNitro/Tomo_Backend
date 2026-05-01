package middleware

import (
	"net/http"
	"strings"

	"example.com/tomo/internal/helper"
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
			tokenClaims, err := helper.ParseTokenClaims(claims)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}

			c.Set(helper.ContextActorID, tokenClaims.Subject)
			c.Set(helper.ContextActorType, tokenClaims.ActorType)
			c.Set(helper.ContextParentID, tokenClaims.ParentID)
			return next(c)
		}
	}
}

func ParentOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if helper.GetActorType(c) != helper.ActorTypeParent {
				return echo.NewHTTPError(http.StatusForbidden, "Parent access required")
			}

			return next(c)
		}
	}
}

func ChildOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if helper.GetActorType(c) != helper.ActorTypeChild {
				return echo.NewHTTPError(http.StatusForbidden, "Child access required")
			}

			return next(c)
		}
	}
}
