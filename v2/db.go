package main

import (
	"fmt"

	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

func (g *GCommited) options() string {
	if g.TwoThiefVariant {
		return "Two Thief Variant"
	}
	return ""
}

func (g *GCommited) UndoKey(c *gin.Context) string {
	cu, err := user.FromSession(c)
	if err != nil || cu == nil || g == nil || g.Key == nil {
		return ""
	}
	return fmt.Sprintf("%s/uid-%d", g.Key, cu.ID())
}
