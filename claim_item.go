package got

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/restful"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(claimItemEntry))
}

func (client *Client) claimItem() {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	g := client.Game
	cp := g.CurrentPlayer()
	g.Phase = claimItem
	e := g.newClaimItemEntryFor(cp)
	restful.AddNoticef(client.Context, string(e.HTML(g)))

	card := g.SelectedThiefArea().Card
	g.SelectedThiefArea().Card = nil
	g.SelectedThiefArea().Thief = noPID

	switch {
	case g.Turn == 1:
		card.FaceUp = true
		cp.Hand.append(card)
		client.drawCard()
	case g.Stepped == 1:
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
		g.SelectedThiefAreaF = g.SelectedAreaF
		g.ClickAreas = nil
		client.startMoveThief()
	default:
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
		client.drawCard()
	}
}

type claimItemEntry struct {
	*Entry
	Area Area
}

func (g *Game) newClaimItemEntryFor(p *Player) *claimItemEntry {
	e := &claimItemEntry{
		Entry: g.newEntryFor(p),
		Area:  *(g.SelectedThiefArea()),
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return e
}

func (e *claimItemEntry) HTML(g *Game) template.HTML {
	return restful.HTML("%s claimed %s card at %s%s.",
		g.NameByPID(e.PlayerID), e.Area.Card.Type, e.Area.RowString(), e.Area.ColString())
}

func (g *Game) finalClaim(c *gin.Context) {
	g.Phase = finalClaim
	for _, row := range g.Grid {
		for _, g.SelectedThiefAreaF = range row {
			if p := g.PlayerByID(g.SelectedThiefAreaF.Thief); p != nil {
				card := g.SelectedThiefAreaF.Card
				g.SelectedThiefAreaF.Card = nil
				g.SelectedThiefAreaF.Thief = noPID
				p.DiscardPile = append(Cards{card}, p.DiscardPile...)
			}
		}
	}
	for _, p := range g.Players() {
		p.Hand.append(p.DiscardPile...)
		p.Hand.append(p.DrawPile...)
		for _, card := range p.Hand {
			card.FaceUp = true
		}
		p.DiscardPile, p.DrawPile = make(Cards, 0), make(Cards, 0)
	}
}
