package main

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(moveThiefEntry))
}

func (g *History) startMoveThief(c *gin.Context) {
	g.Phase = moveThief
	g.ClickAreas = nil
}

func (client Client) moveThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := client.getHistory(c)
	if err != nil {
		jerr(c, err)
		return
	}

	err = g.moveThief(c)
	if err != nil {
		jerr(c, err)
		return
	}

	ks, es := g.cache()
	log.Debugf("ks: %v", ks)
	_, err = client.DS.Put(c, ks, es)
	if err != nil {
		jerr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *History) moveThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	sa, ta, err := g.validateMoveThief(c)
	if err != nil {
		return err
	}

	cp := g.CurrentPlayer()
	// e := g.newMoveThiefEntryFor(cp)
	// restful.AddNoticef(c, string(e.HTML(g)))

	switch {
	case g.PlayedCard.Type == sword:
		g.BumpedPlayerID = sa.Thief
		bumpedTo := g.bumpedTo(ta, sa)
		bumpedTo.Thief = g.BumpedPlayerID
		g.BumpedPlayer().Score += bumpedTo.Card.Value() - sa.Card.Value()
		g.claimItem(c)
	case g.PlayedCard.Type == turban && g.Stepped == 0:
		g.Stepped = 1
	case g.PlayedCard.Type == turban && g.Stepped == 1:
		g.Stepped = 2
		g.claimItem(c)
	default:
		g.claimItem(c)
	}
	sa.Thief = cp.ID
	cp.Score += sa.Card.Value()
	return nil
}

func (g *History) validateMoveThief(c *gin.Context) (*Area, *Area, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	sa, err := g.getAreaFrom(c)
	if err != nil {
		return nil, nil, err
	}

	ta := g.SelectedThiefArea()
	err = g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, nil, err
	case sa == nil:
		return nil, nil, fmt.Errorf("you must select a space to which to move your thief: %w", sn.ErrValidation)
	case ta == nil:
		return nil, nil, fmt.Errorf("thief not selected: %w", sn.ErrValidation)
	case ta.Thief != g.CurrentPlayer().ID:
		return nil, nil, fmt.Errorf("you must first select one of your thieves: %w", sn.ErrValidation)
	case (g.PlayedCard.Type == lamp || g.PlayedCard.Type == sLamp) && !hasArea(g.lampAreas(ta), sa):
		return nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case (g.PlayedCard.Type == camel || g.PlayedCard.Type == sCamel) && !hasArea(g.camelAreas(ta), sa):
		return nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Type == coins && !g.isCoinsArea(sa):
		return nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Type == sword && !g.isSwordArea(sa):
		return nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Type == carpet && !g.isCarpetArea(sa):
		return nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Type == turban && g.Stepped == 0 && !g.isTurban0Area(sa):
		return nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Type == turban && g.Stepped == 1 && !g.isTurban1Area(sa):
		return nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Type == guard:
		return nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	default:
		return sa, ta, nil
	}
}

type moveThiefEntry struct {
	*Entry
	Card Card
	From Area
	To   Area
}

//func (g *History) newMoveThiefEntryFor(p *Player) (e *moveThiefEntry) {
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
//func (e *moveThiefEntry) HTML(g *History) (t template.HTML) {
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

func (g *History) bumpedTo(from, to *Area) *Area {
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
