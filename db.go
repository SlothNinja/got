package got

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	gtype "github.com/SlothNinja/type"
	"github.com/gin-gonic/gin"
)

const kind = "Game"

// New creates a new Guild of Thieves game.
func New(c *gin.Context, id int64) *Game {
	g := new(Game)
	g.Header = game.NewHeader(c, g, id)
	g.State = newState()
	g.Key.Parent = pk(c)
	g.Type = gtype.GOT
	return g
}

func newState() *State {
	return &State{TempData: new(TempData)}
}

func pk(c *gin.Context) *datastore.Key {
	return datastore.NameKey(gtype.GOT.SString(), "root", game.GamesRoot(c))
}

func (client Client) init(c *gin.Context, g *Game) error {
	err := client.Game.AfterLoad(c, g.Header)
	if err != nil {
		return err
	}
	for _, player := range g.Players() {
		player.init(g)
	}
	return nil
}

func (client Client) afterCache(c *gin.Context, g *Game) error {
	return client.init(c, g)
}

func (g *Game) options() (s string) {
	if g.TwoThiefVariant {
		s = "Two Thief Variant"
	}
	return
}

func (g *Game) fromForm(c *gin.Context) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s := struct {
		TwoThiefVariant bool `form:"two-thief-variant"`
	}{}
	err := c.ShouldBind(&s)
	if err != nil {
		return err
	}
	g.TwoThiefVariant = s.TwoThiefVariant
	return nil
}
