package main

import (
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func (g *Game) claimItem(cp *Player) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	sa := g.SelectedThiefArea()

	loggedArea := *sa
	card := sa.Card

	sa.Card = nil
	sa.Thief = noPID

	switch {
	case g.Turn == 4:
		g.appendEntry(Message{
			"template": "claim-item",
			"area":     loggedArea,
			"hand":     true,
		})
		card.FaceUp = true
		cp.Hand.append(card)
	case g.Stepped == 1:
		g.appendEntry(Message{
			"template": "claim-item",
			"area":     loggedArea,
			"hand":     false,
		})
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
	default:
		g.appendEntry(Message{
			"template": "claim-item",
			"area":     loggedArea,
			"hand":     false,
		})
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
		g.drawCard(cp)
	}
}

func (g *Game) finalClaim(c *gin.Context) {
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
