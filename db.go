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
	return new(State)
}

func pk(c *gin.Context) *datastore.Key {
	return datastore.NameKey(gtype.GOT.SString(), "root", game.GamesRoot(c))
}

func (g *Game) init() {
	g.Header.AfterLoad()

	for _, p := range g.Players() {
		p.init(g)
	}
}

func (g *Game) afterCache() {
	g.init()
}

func (g *Game) options() string {
	if g.TwoThiefVariant {
		return "Two Thief Variant"
	}
	return ""
}

func (g *Game) fromForm(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

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
