package got

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/restful"
)

func init() {
	gob.Register(new(drawCardEntry))
}

func (client *Client) drawCard() {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	client.Game.Phase = drawCard
	cp := client.Game.CurrentPlayer()

	if client.Game.Turn != 1 {
		card, shuffle := cp.draw()
		e := client.newDrawCardEntryFor(cp, card, shuffle)
		restful.AddNoticef(client.Context, string(e.HTML(client.Game)))
		if client.Game.PlayedCard.Type == coins {
			card, shuffle := cp.draw()
			e := client.newDrawCardEntryFor(cp, card, shuffle)
			restful.AddNoticef(client.Context, string(e.HTML(client.Game)))
		}
	}
	cp.PerformedAction = true
	client.html("got/move_thief_update")
}

type drawCardEntry struct {
	*Entry
	Card    Card
	Shuffle bool
}

func (client *Client) newDrawCardEntryFor(p *Player, c *Card, shuffle bool) *drawCardEntry {
	e := &drawCardEntry{
		Entry:   client.newEntryFor(p),
		Card:    *c,
		Shuffle: shuffle,
	}
	p.Log = append(p.Log, e)
	client.Game.Log = append(client.Game.Log, e)
	return e
}

func (e *drawCardEntry) HTML(g *Game) (t template.HTML) {
	n := g.NameByPID(e.PlayerID)
	if e.Shuffle {
		t = restful.HTML("%s shuffled discard pile and drew card from newly formed draw pile.", n)
	} else {
		t = restful.HTML("%s drew card from draw pile.", n)
	}
	return
}

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
