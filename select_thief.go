package got

import (
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (g *Game) startSelectThief(c *gin.Context) (tmpl string, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = selectThief
	return "got/played_card_update", nil
}

func (g *Game) selectThief(c *gin.Context) (string, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := g.validateSelectThief(c)
	if err != nil {
		return "got/flash_notice", err
	}

	g.SelectedThiefAreaF = g.SelectedArea()
	return g.startMoveThief(c)
}

func (g *Game) validateSelectThief(c *gin.Context) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	switch area, err := g.SelectedArea(), g.validatePlayerAction(c); {
	case err != nil:
		return err
	case area == nil || area.Thief != g.CurrentPlayer().ID():
		return sn.NewVError("You must select one of your thieves.")
	default:
		return nil
	}
}
