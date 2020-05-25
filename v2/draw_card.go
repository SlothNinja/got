package main

func (g *Game) drawCard(cp *Player) {
	if g.Turn != 1 {
		_, shuffle := cp.draw()
		g.appendEntry(Message{
			"template": "draw-card",
			"shuffled": shuffle,
		})
		if g.PlayedCard.Kind == coinsCard {
			_, shuffle = cp.draw()
			g.appendEntry(Message{
				"template": "draw-card",
				"shuffled": shuffle,
			})
		}
	}
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
