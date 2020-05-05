package main

import (
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func (g *Game) claimItem(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := g.CurrentPlayer()
	g.Phase = claimItem
	// e := g.newClaimItemEntryFor(cp)
	// restful.AddNoticef(c, string(e.HTML(g)))

	card := g.SelectedThiefArea().Card
	g.SelectedThiefArea().Card = nil
	g.SelectedThiefArea().Thief = noPID
	switch {
	case g.Turn == 1:
		card.FaceUp = true
		cp.Hand.append(card)
		g.drawCard(c)
	case g.Stepped == 1:
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
		g.SelectedThiefAreaID = g.SelectedAreaID
		g.ClickAreas = nil
	default:
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
	}
}

// type claimItemEntry struct {
// 	*Entry
// 	Area Area
// }
//
// func (g *Game) newClaimItemEntryFor(p *Player) *claimItemEntry {
// 	e := &claimItemEntry{
// 		Entry: g.newEntryFor(p),
// 		Area:  *(g.SelectedThiefArea()),
// 	}
// 	p.Log = append(p.Log, e)
// 	g.Log = append(g.Log, e)
// 	return e
// }

// func (e *claimItemEntry) HTML(g *Game) template.HTML {
// 	return restful.HTML("%s claimed %s card at %s%s.",
// 		g.NameByPID(e.PlayerID), e.Area.Card.Type, e.Area.RowString(), e.Area.ColString())
// }

func (g *Game) finalClaim(c *gin.Context) {
	g.Phase = finalClaim
	for _, row := range g.Grid {
		for _, a := range row {
			if p := g.PlayerByID(a.Thief); p != nil {
				card := a.Card
				a.Card = nil
				a.Thief = noPID
				p.DiscardPile = append(Cards{card}, p.DiscardPile...)
			}
		}
	}
	for _, p := range g.Players {
		p.Hand.append(p.DiscardPile...)
		p.Hand.append(p.DrawPile...)
		for _, card := range p.Hand {
			card.FaceUp = true
		}
		p.DiscardPile, p.DrawPile = make(Cards, 0), make(Cards, 0)
	}
}
