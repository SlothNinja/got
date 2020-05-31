package main

func (g *game) drawCard(cp *player) {
	if g.Turn != 1 {
		_, shuffle := cp.draw()
		g.appendEntry(message{
			"template": "draw-card",
			"shuffled": shuffle,
		})
		if g.playedCard.Kind == coinsCard {
			_, shuffle = cp.draw()
			g.appendEntry(message{
				"template": "draw-card",
				"shuffled": shuffle,
			})
		}
	}
}

func (p *player) draw() (*Card, bool) {
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
