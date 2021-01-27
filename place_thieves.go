package got

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(placeThiefEntry))
}

func (g *Game) placeThieves(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = placeThieves
	return nil
}

func (client *Client) placeThief() {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.validatePlaceThief()
	if err != nil {
		client.flashError(err)
		return
	}

	g := client.Game
	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	cp.Score += g.SelectedArea().Card.Value()
	g.SelectedArea().Thief = cp.ID()

	// Log placement
	e := g.newPlaceThiefEntryFor(cp)
	restful.AddNoticef(client.Context, string(e.HTML(g)))
	client.html("got/place_thief_update")
}

func (client *Client) validatePlaceThief() error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.validatePlayerAction()
	if err != nil {
		return err
	}

	//g.debugf("Place Thief Area: %#v", g.SelectedArea)

	switch area := client.Game.SelectedArea(); {
	case area == nil:
		return sn.NewVError("You must select an area.")
	case area.Card == nil:
		return sn.NewVError("You must select an area with a card.")
	case area.Thief != noPID:
		return sn.NewVError("You must select an area without a thief.")
	default:
		return nil
	}
}

type placeThiefEntry struct {
	*Entry
	Area Area
}

func (g *Game) newPlaceThiefEntryFor(p *Player) (e *placeThiefEntry) {
	area := g.SelectedArea()
	e = &placeThiefEntry{
		Entry: g.newEntryFor(p),
		Area:  *area,
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *placeThiefEntry) HTML(g *Game) template.HTML {
	return restful.HTML("%s placed thief on %s at %s%s.",
		g.NameByPID(e.PlayerID), e.Area.Card.Type, e.Area.RowString(), e.Area.ColString())
}
