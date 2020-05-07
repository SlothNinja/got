package main

import (
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func (h *History) claimItem(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := h.CurrentPlayer()
	h.Phase = claimItem
	// e := g.newClaimItemEntryFor(cp)
	// restful.AddNoticef(c, string(e.HTML(g)))

	card := h.SelectedThiefArea().Card
	h.SelectedThiefArea().Card = nil
	h.SelectedThiefArea().Thief = noPID
	switch {
	case h.Turn == 1:
		card.FaceUp = true
		cp.Hand.append(card)
	case h.Stepped == 1:
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
		h.SelectedThiefAreaID = h.SelectedAreaID
		h.ClickAreas = nil
	default:
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
		h.drawCard(c)
	}
}

// type claimItemEntry struct {
// 	*Entry
// 	Area Area
// }
//
// func (g *History) newClaimItemEntryFor(p *Player) *claimItemEntry {
// 	e := &claimItemEntry{
// 		Entry: g.newEntryFor(p),
// 		Area:  *(g.SelectedThiefArea()),
// 	}
// 	p.Log = append(p.Log, e)
// 	g.Log = append(g.Log, e)
// 	return e
// }

// func (e *claimItemEntry) HTML(g *History) template.HTML {
// 	return restful.HTML("%s claimed %s card at %s%s.",
// 		g.NameByPID(e.PlayerID), e.Area.Card.Type, e.Area.RowString(), e.Area.ColString())
// }

func (g *History) finalClaim(c *gin.Context) {
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
