package helper

import (
	"github.com/labstack/echo/v5"
)

const (
	ContextActorID   = "actor_id"
	ContextActorType = "actor_type"
	ContextParentID  = "parent_id"
)

func GetUserID(c *echo.Context) string {
	return GetActorID(c)
}

func GetActorID(c *echo.Context) string {
	actorID, ok := c.Get(ContextActorID).(string)
	if !ok || actorID == "" {
		return ""
	}

	return actorID
}

func GetActorType(c *echo.Context) string {
	actorType, ok := c.Get(ContextActorType).(string)
	if !ok || actorType == "" {
		return ""
	}

	return actorType
}

func GetParentID(c *echo.Context) string {
	parentID, ok := c.Get(ContextParentID).(string)
	if !ok || parentID == "" {
		return ""
	}

	return parentID
}
