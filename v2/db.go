package main

import (
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/sn/v2"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

// newGame creates a new Guild of Thieves game.
func newGame(id int64) *Game {
	g := new(Game)
	g.Key = newGameKey(id)
	g.Type = sn.GOT
	return g
}

func newGameKey(id int64) *datastore.Key {
	return datastore.IDKey(gameKind, id, rootKey(id))
}

func (g *Game) options() string {
	if g.TwoThiefVariant {
		return "Two Thief Variant"
	}
	return ""
}

func (g *Game) UndoKey(c *gin.Context) string {
	cu, err := user.FromSession(c)
	if err != nil || cu == nil || g == nil || g.Key == nil {
		return ""
	}
	return fmt.Sprintf("%s/uid-%d", g.Key, cu.ID())
}
