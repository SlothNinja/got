package got

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(passEntry))
}

func (g *Game) pass(c *gin.Context) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if err := g.validatePass(c); err != nil {
		return "got/flash_notice", game.None, err
	}

	cp := g.CurrentPlayer()
	cp.Passed = true
	cp.PerformedAction = true
	g.Phase = drawCard

	// Log Pass
	e := g.newPassEntryFor(cp)
	restful.AddNoticef(c, string(e.HTML(g)))

	return "got/pass_update", game.Cache, nil
}

func (g *Game) validatePass(c *gin.Context) error {
	if err := g.validatePlayerAction(c); err != nil {
		return err
	}
	return nil
}

type passEntry struct {
	*Entry
}

func (g *Game) newPassEntryFor(p *Player) (e *passEntry) {
	e = &passEntry{
		Entry: g.newEntryFor(p),
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *passEntry) HTML(g *Game) template.HTML {
	return restful.HTML("%s passed.", g.NameByPID(e.PlayerID))
}
