package got

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	gtype "github.com/SlothNinja/type"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

func (g *Game) init(c *gin.Context) (err error) {
	if err = g.Header.AfterLoad(g); err == nil {
		for _, player := range g.Players() {
			player.init(g)
		}
	}

	//	for _, entry := range g.Log {
	//		entry.Init(g)
	//	}
	return
}

func (g *Game) afterCache() error {
	return g.init(g.CTX())
}

func (g *Game) options() (s string) {
	if g.TwoThiefVariant {
		s = "Two Thief Variant"
	}
	return
}

func (g *Game) fromForm(c *gin.Context) (err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	s := new(State)
	if err = restful.BindWith(c, s, binding.FormPost); err == nil {
		g.TwoThiefVariant = s.TwoThiefVariant
	}
	log.Debugf("err: %v s:%#v", err, s)
	return
}
