package got

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
)

func init() {
	gob.Register(new(placeThiefEntry))
}

func (client *Client) placeThieves() {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	client.Game.Phase = placeThieves
}

func (client *Client) placeThief() {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.validatePlaceThief()
	if err != nil {
		client.flashError(err)
		return
	}

	cp := client.Game.CurrentPlayer()
	cp.PerformedAction = true
	cp.Score += client.Game.SelectedArea().Card.Value()
	client.Game.SelectedArea().Thief = cp.ID()

	// Log placement
	e := client.newPlaceThiefEntryFor(cp)
	restful.AddNoticef(client.Context, string(e.HTML(client.Game)))
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

func (client *Client) newPlaceThiefEntryFor(p *Player) *placeThiefEntry {
	area := client.Game.SelectedArea()
	e := &placeThiefEntry{
		Entry: client.newEntryFor(p),
		Area:  *area,
	}
	p.Log = append(p.Log, e)
	client.Game.Log = append(client.Game.Log, e)
	return e
}

func (e *placeThiefEntry) HTML(g *Game) template.HTML {
	return restful.HTML("%s placed thief on %s at %s%s.",
		g.NameByPID(e.PlayerID), e.Area.Card.Type, e.Area.RowString(), e.Area.ColString())
}
