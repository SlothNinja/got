package main

import (
	"github.com/gin-gonic/gin"
)

func (h *History) drawCard(c *gin.Context) {
	h.Phase = drawCard
	cp := h.CurrentPlayer()

	if h.Turn != 1 {
		cp.draw()
		// card, shuffle := cp.draw()
		// e := g.newDrawCardEntryFor(cp, card, shuffle)
		// restful.AddNoticef(c, string(e.HTML(g)))
		if h.PlayedCard.Type == coins {
			cp.draw()
			// card, shuffle := cp.draw()
			// e := g.newDrawCardEntryFor(cp, card, shuffle)
			// restful.AddNoticef(c, string(e.HTML(g)))
		}
	}
	cp.PerformedAction = true
}

// type drawCardEntry struct {
// 	*Entry
// 	Card    Card
// 	Shuffle bool
// }
//
// func (g *History) newDrawCardEntryFor(p *Player, c *Card, shuffle bool) *drawCardEntry {
// 	e := &drawCardEntry{
// 		Entry:   g.newEntryFor(p),
// 		Card:    *c,
// 		Shuffle: shuffle,
// 	}
// 	p.Log = append(p.Log, e)
// 	g.Log = append(g.Log, e)
// 	return e
// }
//
// func (e *drawCardEntry) HTML(g *History) (t template.HTML) {
// 	n := g.NameByPID(e.PlayerID)
// 	if e.Shuffle {
// 		t = restful.HTML("%s shuffled discard pile and drew card from newly formed draw pile.", n)
// 	} else {
// 		t = restful.HTML("%s drew card from draw pile.", n)
// 	}
// 	return
// }

func (p *Player) draw() (*Card, bool) {
	shuffle := false
	if len(p.DrawPile) == 0 {
		shuffle = true
		p.DrawPile = p.DiscardPile
		for _, card := range p.DrawPile {
			card.FaceUp = false
		}
		p.DiscardPile = make(Cards, 0)
	}
	card := p.DrawPile.draw()
	p.Hand.append(card)
	return card, shuffle
}
