package main

import (
	"encoding/gob"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(moveThiefEntry))
}

func (g *Game) startMoveThief(c *gin.Context) {
	g.Phase = moveThief
	g.ClickAreas = nil
}

func (g *Game) moveThief(c *gin.Context) error {
	err := g.validateMoveThief(c)
	if err != nil {
		return err
	}

	cp := g.CurrentPlayer()
	// e := g.newMoveThiefEntryFor(cp)
	// restful.AddNoticef(c, string(e.HTML(g)))

	switch {
	case g.PlayedCard.Type == sword:
		g.BumpedPlayerID = g.SelectedArea().Thief
		bumpedTo := g.bumpedTo(g.SelectedThiefArea(), g.SelectedArea())
		bumpedTo.Thief = g.BumpedPlayerID
		g.BumpedPlayer().Score += bumpedTo.Card.Value() - g.SelectedArea().Card.Value()
	case g.PlayedCard.Type == turban && g.Stepped == 0:
		g.Stepped = 1
	case g.PlayedCard.Type == turban && g.Stepped == 1:
		g.Stepped = 2
	}
	g.SelectedArea().Thief = cp.ID
	cp.Score += g.SelectedArea().Card.Value()
	g.claimItem(c)
	return nil
}

func (g *Game) validateMoveThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	a := g.SelectedArea()
	g.ClickAreas = nil
	err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return err
	case a == nil:
		return sn.NewVError("You must select a space which to move your thief.")
	case g.SelectedThiefArea() != nil && g.SelectedThiefArea().Thief != g.CurrentPlayer().ID:
		return sn.NewVError("You must first select one of your thieves.")
	case (g.PlayedCard.Type == lamp || g.PlayedCard.Type == sLamp) && !g.isLampArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case (g.PlayedCard.Type == camel || g.PlayedCard.Type == sCamel) && !g.isCamelArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case g.PlayedCard.Type == coins && !g.isCoinsArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case g.PlayedCard.Type == sword && !g.isSwordArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case g.PlayedCard.Type == carpet && !g.isCarpetArea(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case g.PlayedCard.Type == turban && g.Stepped == 0 && !g.isTurban0Area(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case g.PlayedCard.Type == turban && g.Stepped == 1 && !g.isTurban1Area(a):
		return sn.NewVError("You can't move the selected thief to area %d%d", a.Row, a.Column)
	case g.PlayedCard.Type == guard:
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

//func (g *Game) newMoveThiefEntryFor(p *Player) (e *moveThiefEntry) {
//	e = &moveThiefEntry{
//		Entry: g.newEntryFor(p),
//		Card:  *(g.PlayedCard),
//		From:  *(g.SelectedThiefArea()),
//		To:    *(g.SelectedArea()),
//	}
//	if g.JewelsPlayed {
//		e.Card = *(newCard(jewels, true))
//	}
//	p.Log = append(p.Log, e)
//	g.Log = append(g.Log, e)
//	return
//}
//
//func (e *moveThiefEntry) HTML(g *Game) (t template.HTML) {
//	from := e.From
//	to := e.To
//	n := g.NameByPID(e.PlayerID)
//	if e.Card.Type == sword {
//		bumped := g.bumpedTo(&from, &to)
//		t = restful.HTML("%s moved thief from %s card at %s%s to %s card at %s%s and bumped thief to card at %s%s.",
//			n, from.Card.Type, from.RowString(), from.ColString(), to.Card.Type,
//			to.RowString(), to.ColString(), bumped.RowString(), bumped.ColString())
//	} else {
//		t = restful.HTML("%s moved thief from %s card at %s%s to %s card at %s%s.", n,
//			from.Card.Type, from.RowString(), from.ColString(), to.Card.Type, to.RowString(),
//			to.ColString())
//	}
//	return
//}

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