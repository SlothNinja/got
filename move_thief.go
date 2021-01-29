package got

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
)

func init() {
	gob.Register(new(moveThiefEntry))
}

func (client *Client) startMoveThief() {
	client.Game.Phase = moveThief
	client.Game.ClickAreas = nil
	client.html("got/select_thief_update")
}

func (client *Client) moveThief() {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.validateMoveThief()
	if err != nil {
		client.flashError(err)
		return
	}

	cp := client.Game.CurrentPlayer()
	e := client.newMoveThiefEntryFor(cp)
	restful.AddNoticef(client.Context, string(e.HTML(client.Game)))

	switch {
	case client.Game.PlayedCard.Type == sword:
		client.Game.BumpedPlayerID = client.Game.SelectedArea().Thief
		bumpedTo := client.Game.bumpedTo(client.Game.SelectedThiefArea(), client.Game.SelectedArea())
		bumpedTo.Thief = client.Game.BumpedPlayerID
		client.Game.BumpedPlayer().Score += bumpedTo.Card.Value() - client.Game.SelectedArea().Card.Value()
	case client.Game.PlayedCard.Type == turban && client.Game.Stepped == 0:
		client.Game.Stepped = 1
	case client.Game.PlayedCard.Type == turban && client.Game.Stepped == 1:
		client.Game.Stepped = 2
	}
	client.Game.SelectedArea().Thief = cp.ID()
	cp.Score += client.Game.SelectedArea().Card.Value()
	client.claimItem()
}

func (client *Client) validateMoveThief() error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	a := client.Game.SelectedArea()
	client.Game.ClickAreas = nil

	err := client.validatePlayerAction()
	switch {
	case err != nil:
		return err
	case a == nil:
		return sn.NewVError("You must select a space which to move your thief.")
	case client.Game.SelectedThiefArea() != nil && client.Game.SelectedThiefArea().Thief != client.Game.CurrentPlayer().ID():
		return sn.NewVError("You must first select one of your thieves.")
	case (client.Game.PlayedCard.Type == lamp || client.Game.PlayedCard.Type == sLamp) && !client.Game.isLampArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case (client.Game.PlayedCard.Type == camel || client.Game.PlayedCard.Type == sCamel) && !client.Game.isCamelArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case client.Game.PlayedCard.Type == coins && !client.Game.isCoinsArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case client.Game.PlayedCard.Type == sword && !client.Game.isSwordArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case client.Game.PlayedCard.Type == carpet && !client.Game.isCarpetArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case client.Game.PlayedCard.Type == turban && client.Game.Stepped == 0 && !client.Game.isTurban0Area(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case client.Game.PlayedCard.Type == turban && client.Game.Stepped == 1 && !client.Game.isTurban1Area(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case client.Game.PlayedCard.Type == guard:
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	default:
		return nil
	}
}

type moveThiefEntry struct {
	*Entry
	Card Card
	From Area
	To   Area
}

func (client *Client) newMoveThiefEntryFor(p *Player) *moveThiefEntry {
	e := &moveThiefEntry{
		Entry: client.newEntryFor(p),
		Card:  *(client.Game.PlayedCard),
		From:  *(client.Game.SelectedThiefArea()),
		To:    *(client.Game.SelectedArea()),
	}
	if client.Game.JewelsPlayed {
		e.Card = *(newCard(jewels, true))
	}
	p.Log = append(p.Log, e)
	client.Game.Log = append(client.Game.Log, e)
	return e
}

func (e *moveThiefEntry) HTML(g *Game) (t template.HTML) {
	from := e.From
	to := e.To
	n := g.NameByPID(e.PlayerID)
	if e.Card.Type == sword {
		bumped := g.bumpedTo(&from, &to)
		t = restful.HTML("%s moved thief from %s card at %s%s to %s card at %s%s and bumped thief to card at %s%s.",
			n, from.Card.Type, from.RowString(), from.ColString(), to.Card.Type,
			to.RowString(), to.ColString(), bumped.RowString(), bumped.ColString())
	} else {
		t = restful.HTML("%s moved thief from %s card at %s%s to %s card at %s%s.", n,
			from.Card.Type, from.RowString(), from.ColString(), to.Card.Type, to.RowString(),
			to.ColString())
	}
	return
}

func (g *Game) bumpedTo(from, to *Area) *Area {
	switch {
	case from.Row > to.Row:
		return g.Grid[to.Row-1][from.Column]
	case from.Row < to.Row:
		return g.Grid[to.Row+1][from.Column]
	case from.Column > to.Column:
		return g.Grid[from.Row][to.Column-1]
	case from.Column < to.Column:
		return g.Grid[from.Row][to.Column+1]
	default:
		return nil
	}
}
