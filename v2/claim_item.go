package main

import (
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func (g *Game) claimItem(cp *player, a *Area) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	loggedArea := *a
	card := a.Card

	a.Card = nil
	a.Thief = noPID

	switch {
	case g.Turn == 4:
		g.appendEntry(message{
			"template": "claim-item",
			"area":     loggedArea,
			"hand":     true,
		})
		card.FaceUp = true
		cp.Hand.append(card)
		cp.Stats.Claimed.inc(card.Kind)
	case g.stepped == 1:
		g.appendEntry(message{
			"template": "claim-item",
			"area":     loggedArea,
			"hand":     false,
		})
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
		cp.Stats.Claimed.inc(card.Kind)
	default:
		g.appendEntry(message{
			"template": "claim-item",
			"area":     loggedArea,
			"hand":     false,
		})
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
		cp.Stats.Claimed.inc(card.Kind)
		g.drawCard(cp)
	}
}

func (g *Game) finalClaim(c *gin.Context) {
	for _, row := range g.grid {
		for _, a := range row {
			if p := g.playerByID(a.Thief); p != nil {
				card := a.Card
				a.Card = nil
				a.Thief = noPID
				p.DiscardPile = append(Cards{card}, p.DiscardPile...)
				p.Stats.Claimed.inc(card.Kind)
			}
		}
	}
	for _, p := range g.players {
		p.Hand.append(p.DiscardPile...)
		p.Hand.append(p.DrawPile...)
		for _, card := range p.Hand {
			card.FaceUp = true
		}
		p.DiscardPile, p.DrawPile = make(Cards, 0), make(Cards, 0)
	}
}
