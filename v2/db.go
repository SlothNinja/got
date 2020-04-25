package main

import (
	"fmt"

	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

const kind = "Game"

// New creates a new Guild of Thieves game.
func New(c *gin.Context, id int64) *Game {
	g := new(Game)
	g.Header = newHeader(id)
	g.State = new(State)
	g.Type = sn.GOT
	return g
}

func (g *Game) options() string {
	if g.TwoThiefVariant {
		return "Two Thief Variant"
	}
	return ""
}

func (g *Game) UndoKey(c *gin.Context) string {
	cu, err := user.FromSession(c)
	if err != nil || cu == nil || g == nil || g.Header == nil {
		return ""
	}
	return fmt.Sprintf("%s/uid-%d", g.Header.Key, cu.ID())
}
