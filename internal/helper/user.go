package helper

import (
	"github.com/labstack/echo/v5"
)

func GetUserID(c *echo.Context) string {
	userId, ok := c.Get("user_id").(string)
	if !ok || userId == "" {
		return ""
	}
	return userId
}
