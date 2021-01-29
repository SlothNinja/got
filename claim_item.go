package got

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/restful"
)

func init() {
	gob.Register(new(claimItemEntry))
}

func (client *Client) claimItem() {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	cp := client.Game.CurrentPlayer()
	client.Game.Phase = claimItem
	e := client.newClaimItemEntryFor(cp)
	restful.AddNoticef(client.Context, string(e.HTML(client.Game)))

	card := client.Game.SelectedThiefArea().Card
	client.Game.SelectedThiefArea().Card = nil
	client.Game.SelectedThiefArea().Thief = noPID

	switch {
	case client.Game.Turn == 1:
		card.FaceUp = true
		cp.Hand.append(card)
		client.drawCard()
	case client.Game.Stepped == 1:
		cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
		client.Game.SelectedThiefAreaF = client.Game.SelectedAreaF
		client.Game.ClickAreas = nil
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

func (client *Client) newClaimItemEntryFor(p *Player) *claimItemEntry {
	e := &claimItemEntry{
		Entry: client.newEntryFor(p),
		Area:  *(client.Game.SelectedThiefArea()),
	}
	p.Log = append(p.Log, e)
	client.Game.Log = append(client.Game.Log, e)
	return e
}

func (e *claimItemEntry) HTML(g *Game) template.HTML {
	return restful.HTML("%s claimed %s card at %s%s.",
		g.NameByPID(e.PlayerID), e.Area.Card.Type, e.Area.RowString(), e.Area.ColString())
}

func (client *Client) finalClaim() {
	client.Game.Phase = finalClaim
	for _, row := range client.Game.Grid {
		for _, client.Game.SelectedThiefAreaF = range row {
			if p := client.Game.PlayerByID(client.Game.SelectedThiefAreaF.Thief); p != nil {
				card := client.Game.SelectedThiefAreaF.Card
				client.Game.SelectedThiefAreaF.Card = nil
				client.Game.SelectedThiefAreaF.Thief = noPID
				p.DiscardPile = append(Cards{card}, p.DiscardPile...)
			}
		}
	}
	for _, p := range client.Game.Players() {
		p.Hand.append(p.DiscardPile...)
		p.Hand.append(p.DrawPile...)
		for _, card := range p.Hand {
			card.FaceUp = true
		}
		p.DiscardPile, p.DrawPile = make(Cards, 0), make(Cards, 0)
	}
}
