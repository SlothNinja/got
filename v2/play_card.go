package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (cl client) playCard(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cp, err := g.playCard(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	ks, es := g.cache()
	_, err = cl.DS.Put(c, ks, es)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	g.updateClickablesFor(c, cp, g.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *game) playCard(c *gin.Context) (*player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	// reset card played related flags
	g.stepped = 0
	g.playedCard = nil

	cp, card, err := g.validatePlayCard(c)
	if err != nil {
		return nil, err
	}

	cp.Hand.play(card)
	cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)
	cp.Stats.Played.inc(card.Kind)

	if card.Kind == jewels {
		pc := g.jewels
		cp.Stats.JewelsAs.inc(pc.Kind)
		g.playedCard = &pc
	} else {
		g.playedCard = card
	}

	g.Phase = selectThiefPhase
	g.Undo.Update()

	g.newEntryFor(cp.ID, message{
		"template": "play-card",
		"card":     *card,
	})
	return cp, nil
}

func (g *game) validatePlayCard(c *gin.Context) (*player, *Card, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validatePlayerAction(c)
	if err != nil {
		return nil, nil, err
	}

	card, err := g.getCardFrom(c, cp)
	switch {
	case err != nil:
		return nil, nil, err
	case card == nil:
		return nil, nil, fmt.Errorf("you must select a card: %w", sn.ErrValidation)
	case g.Phase != playCardPhase:
		return nil, nil, fmt.Errorf("expected %q phase but have %q phase: %w",
			playCardPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, card, nil
	}
}

func (g *game) getCardFrom(c *gin.Context, cp *player) (*Card, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	obj := struct {
		Kind cKind `json:"kind"`
	}{}

	err := c.Bind(&obj)
	if err != nil {
		return nil, err
	}

	i, found := cp.Hand.indexFor(newCard(obj.Kind, false))
	if !found {
		return nil, fmt.Errorf("unable to find card: %w", sn.ErrValidation)
	}
	return cp.Hand[i], nil
}

func (g *game) lampAreas(thiefArea *Area) []*Area {
	var as []*Area

	// Move Left
	var a2 *Area
	for col := thiefArea.Column - 1; col >= col1; col-- {
		temp := g.getArea(areaID{thiefArea.Row, col})
		if !canMoveTo(temp) {
			break
		}
		a2 = temp
	}
	if a2 != nil {
		as = append(as, a2)
	}

	// Move right
	a2 = nil
	for col := thiefArea.Column + 1; col <= col8; col++ {
		temp := g.getArea(areaID{thiefArea.Row, col})
		if !canMoveTo(temp) {
			break
		}
		a2 = temp
	}
	if a2 != nil {
		as = append(as, a2)
	}

	// Move Up
	a2 = nil
	for row := thiefArea.Row - 1; row >= rowA; row-- {
		temp := g.getArea(areaID{row, thiefArea.Column})
		if !canMoveTo(temp) {
			break
		}
		a2 = temp
	}
	if a2 != nil {
		as = append(as, a2)
	}

	// Move Down
	a2 = nil
	for row := thiefArea.Row + 1; row <= g.lastRow(); row++ {
		temp := g.getArea(areaID{row, thiefArea.Column})
		if !canMoveTo(temp) {
			break
		}
		a2 = temp
	}
	if a2 != nil {
		as = append(as, a2)
	}
	return as
}

// camelAreas returns areas from thief area ta reachable via a camel card.
func (g *game) camelAreas(ta *Area) []*Area {
	var as []*Area

	// Move Three Left?
	if ta.Column-3 >= col1 {
		area1 := g.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := g.getArea(areaID{ta.Row, ta.Column - 2})
		area3 := g.getArea(areaID{ta.Row, ta.Column - 3})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Three Right?
	if ta.Column+3 <= col8 {
		area1 := g.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := g.getArea(areaID{ta.Row, ta.Column + 2})
		area3 := g.getArea(areaID{ta.Row, ta.Column + 3})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Three Up?
	if ta.Row-3 >= rowA {
		area1 := g.getArea(areaID{ta.Row - 1, ta.Column})
		area2 := g.getArea(areaID{ta.Row - 2, ta.Column})
		area3 := g.getArea(areaID{ta.Row - 3, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Three Down?
	if ta.Row+3 <= g.lastRow() {
		area1 := g.getArea(areaID{ta.Row + 1, ta.Column})
		area2 := g.getArea(areaID{ta.Row + 2, ta.Column})
		area3 := g.getArea(areaID{ta.Row + 3, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Left One Up or One Up Two Left or One Left One Up One Left?
	if ta.Column-2 >= col1 && ta.Row-1 >= rowA {
		area1 := g.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := g.getArea(areaID{ta.Row, ta.Column - 2})
		area3 := g.getArea(areaID{ta.Row - 1, ta.Column - 2})
		area4 := g.getArea(areaID{ta.Row - 1, ta.Column})
		area5 := g.getArea(areaID{ta.Row - 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Left One Down or One Down Two Left or One Left One Down One Left?
	if ta.Column-2 >= col1 && ta.Row+1 <= g.lastRow() {
		area1 := g.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := g.getArea(areaID{ta.Row, ta.Column - 2})
		area3 := g.getArea(areaID{ta.Row + 1, ta.Column - 2})
		area4 := g.getArea(areaID{ta.Row + 1, ta.Column})
		area5 := g.getArea(areaID{ta.Row + 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Right One Up or One Up Two Right or One Right One Up One Right?
	if ta.Column+2 <= col8 && ta.Row-1 >= rowA {
		area1 := g.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := g.getArea(areaID{ta.Row, ta.Column + 2})
		area3 := g.getArea(areaID{ta.Row - 1, ta.Column + 2})
		area4 := g.getArea(areaID{ta.Row - 1, ta.Column})
		area5 := g.getArea(areaID{ta.Row - 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Right One Down or One Down Two Right or One Right One Down One Right?
	if ta.Column+2 <= col8 && ta.Row+1 <= g.lastRow() {
		area1 := g.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := g.getArea(areaID{ta.Row, ta.Column + 2})
		area3 := g.getArea(areaID{ta.Row + 1, ta.Column + 2})
		area4 := g.getArea(areaID{ta.Row + 1, ta.Column})
		area5 := g.getArea(areaID{ta.Row + 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Right Two Down or Two Down One Right or One Down One Right One Down?
	if ta.Column+1 <= col8 && ta.Row+2 <= g.lastRow() {
		area1 := g.getArea(areaID{ta.Row + 1, ta.Column})
		area2 := g.getArea(areaID{ta.Row + 2, ta.Column})
		area3 := g.getArea(areaID{ta.Row + 2, ta.Column + 1})
		area4 := g.getArea(areaID{ta.Row, ta.Column + 1})
		area5 := g.getArea(areaID{ta.Row + 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Right Two Up or Two Up One Right or One Up One Right One Up?
	if ta.Column+1 <= col8 && ta.Row-2 >= rowA {
		area1 := g.getArea(areaID{ta.Row - 1, ta.Column})
		area2 := g.getArea(areaID{ta.Row - 2, ta.Column})
		area3 := g.getArea(areaID{ta.Row - 2, ta.Column + 1})
		area4 := g.getArea(areaID{ta.Row, ta.Column + 1})
		area5 := g.getArea(areaID{ta.Row - 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Left Two Down or Two Down One Left or One Down One Left One Down?
	if ta.Column-1 >= col1 && ta.Row+2 <= g.lastRow() {
		area1 := g.getArea(areaID{ta.Row + 1, ta.Column})
		area2 := g.getArea(areaID{ta.Row + 2, ta.Column})
		area3 := g.getArea(areaID{ta.Row + 2, ta.Column - 1})
		area4 := g.getArea(areaID{ta.Row, ta.Column - 1})
		area5 := g.getArea(areaID{ta.Row + 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Left Two Up or Two Up One Left or One Up One Left One Up?
	if ta.Column-1 >= col1 && ta.Row-2 >= rowA {
		area1 := g.getArea(areaID{ta.Row - 1, ta.Column})
		area2 := g.getArea(areaID{ta.Row - 2, ta.Column})
		area3 := g.getArea(areaID{ta.Row - 2, ta.Column - 1})
		area4 := g.getArea(areaID{ta.Row, ta.Column - 1})
		area5 := g.getArea(areaID{ta.Row - 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Left One Up One Right or One Up One Left One Down?
	if ta.Column-1 >= col1 && ta.Row-1 >= rowA {
		area1 := g.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := g.getArea(areaID{ta.Row - 1, ta.Column - 1})
		area3 := g.getArea(areaID{ta.Row - 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	// Move One Up One Right One Down or One Right One Up One Left?
	if ta.Column+1 <= col8 && ta.Row-1 >= rowA {
		area1 := g.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := g.getArea(areaID{ta.Row - 1, ta.Column + 1})
		area3 := g.getArea(areaID{ta.Row - 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	// Move One Left One Down One Right or One Down One Left One Up?
	if ta.Column-1 >= col1 && ta.Row+1 <= g.lastRow() {
		area1 := g.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := g.getArea(areaID{ta.Row + 1, ta.Column - 1})
		area3 := g.getArea(areaID{ta.Row + 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	// Move One Down One Right One Up or One Right One Down One Left?
	if ta.Column+1 <= col8 && ta.Row+1 <= g.lastRow() {
		area1 := g.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := g.getArea(areaID{ta.Row + 1, ta.Column + 1})
		area3 := g.getArea(areaID{ta.Row + 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	return as
}

func canMoveTo(as ...*Area) bool {
	for _, a := range as {
		if a.hasThief() || !a.hasCard() {
			return false
		}
	}
	return true
}

func (g *game) swordAreas(cp *player, thiefArea *Area) []*Area {
	var as []*Area

	// Move Left
	if area, row := thiefArea, thiefArea.Row; thiefArea.Column >= col3 {
		// Left as far as permitted
		for col := thiefArea.Column - 1; col >= col3; col-- {
			if temp := g.getArea(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := g.getArea(areaID{row, area.Column - 1}), g.getArea(areaID{row, area.Column - 2})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	// Move Right
	if area, row := thiefArea, thiefArea.Row; thiefArea.Column <= col6 {
		// Right as far as permitted
		for col := thiefArea.Column + 1; col <= col6; col++ {
			if temp := g.getArea(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := g.getArea(areaID{row, area.Column + 1}), g.getArea(areaID{row, area.Column + 2})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	// Move Up
	if area, col := thiefArea, thiefArea.Column; thiefArea.Row >= rowC {
		// Up as far as permitted
		for row := thiefArea.Row - 1; row >= rowC; row-- {
			if temp := g.getArea(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := g.getArea(areaID{area.Row - 1, col}), g.getArea(areaID{area.Row - 2, col})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	// Move Down
	if area, col := thiefArea, thiefArea.Column; thiefArea.Row <= g.lastRow()-2 {
		// Down as far as permitted
		for row := thiefArea.Row + 1; row <= g.lastRow()-2; row++ {
			//g.debugf("Row: %v Col: %v", row, col)
			if temp := g.getArea(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := g.getArea(areaID{area.Row + 1, col}), g.getArea(areaID{area.Row + 2, col})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	return as
}

func (p *player) anotherThiefIn(a *Area) bool {
	return a.hasThief() && a.Thief != p.ID
}

func (g *game) isCarpetArea(a *Area) bool {
	if g.selectedThiefArea() != nil {
		return hasArea(g.carpetAreas(), a)
	}
	return false
}

func (g *game) carpetAreas() []*Area {
	as := make([]*Area, 0)
	a1 := g.selectedThiefArea()

	// Move Left
	var a2, empty *Area
MoveLeft:
	for col := a1.Column - 1; col >= col1; col-- {
		switch temp := g.getArea(areaID{a1.Row, col}); {
		case temp.Card == nil:
			empty = temp
		case empty != nil && canMoveTo(temp):
			a2 = temp
			break MoveLeft
		default:
			break MoveLeft
		}
	}
	if a2 != nil {
		as = append(as, a2)
	}

	// Move Right
	a2, empty = nil, nil
MoveRight:
	for col := a1.Column + 1; col <= col8; col++ {
		switch temp := g.getArea(areaID{a1.Row, col}); {
		case temp.Card == nil:
			empty = temp
		case empty != nil && canMoveTo(temp):
			a2 = temp
			break MoveRight
		default:
			break MoveRight
		}
	}
	if a2 != nil {
		as = append(as, a2)
	}

	// Move Up
	a2, empty = nil, nil
MoveUp:
	for row := a1.Row - 1; row >= rowA; row-- {
		switch temp := g.getArea(areaID{row, a1.Column}); {
		case temp.Card == nil:
			empty = temp
		case empty != nil && canMoveTo(temp):
			a2 = temp
			break MoveUp
		default:
			break MoveUp
		}
	}
	if a2 != nil {
		as = append(as, a2)
	}

	// Move Down
	a2, empty = nil, nil
MoveDown:
	for row := a1.Row + 1; row <= g.lastRow(); row++ {
		switch temp := g.getArea(areaID{row, a1.Column}); {
		case temp.Card == nil:
			empty = temp
		case empty != nil && canMoveTo(temp):
			a2 = temp
			break MoveDown
		default:
			break MoveDown
		}
	}
	if a2 != nil {
		as = append(as, a2)
	}

	return as
}

func (g *game) turban0Areas(thiefArea *Area) []*Area {
	var as []*Area

	// Move Left
	if col := thiefArea.Column - 1; col >= col1 {
		if area := g.getArea(areaID{thiefArea.Row, col}); canMoveTo(area) {
			// Left
			if col := col - 1; col >= col1 && canMoveTo(g.getArea(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Up
			if row := area.Row - 1; row >= rowA && canMoveTo(g.getArea(areaID{row, col})) {
				as = append(as, area)
			}
			// Down
			if row := area.Row + 1; row <= g.lastRow() && canMoveTo(g.getArea(areaID{row, col})) {
				as = append(as, area)
			}
		}
	}

	// Move Right
	if col := thiefArea.Column + 1; col <= col8 {
		if area := g.getArea(areaID{thiefArea.Row, col}); canMoveTo(area) {
			// Right
			if col := col + 1; col <= col8 && canMoveTo(g.getArea(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Up
			if row := area.Row - 1; row >= rowA && canMoveTo(g.getArea(areaID{row, col})) {
				as = append(as, area)
			}
			// Down
			if row := area.Row + 1; row <= g.lastRow() && canMoveTo(g.getArea(areaID{row, col})) {
				as = append(as, area)
			}
		}
	}

	// Move Up
	if row := thiefArea.Row - 1; row >= rowA {
		if area := g.getArea(areaID{row, thiefArea.Column}); canMoveTo(area) {
			// Left
			if col := area.Column - 1; col >= col1 && canMoveTo(g.getArea(areaID{row, col})) {
				as = append(as, area)
			}
			// Right
			if col := area.Column + 1; col <= col8 && canMoveTo(g.getArea(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Up
			if row := row - 1; row >= rowA && canMoveTo(g.getArea(areaID{row, thiefArea.Column})) {
				as = append(as, area)
			}
		}
	}

	// Move Down
	if row := thiefArea.Row + 1; row <= g.lastRow() {
		if area := g.getArea(areaID{row, thiefArea.Column}); canMoveTo(area) {
			// Left
			if col := area.Column - 1; col >= col1 && canMoveTo(g.getArea(areaID{row, col})) {
				as = append(as, area)
			}
			// Right
			if col := area.Column + 1; col <= col8 && canMoveTo(g.getArea(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Down
			if row := row + 1; row <= g.lastRow() && canMoveTo(g.getArea(areaID{row, thiefArea.Column})) {
				as = append(as, area)
			}
		}
	}

	return as
}

func (g *game) turban1Areas(thiefArea *Area) []*Area {
	var as []*Area

	// Move Left
	if thiefArea.Column-1 >= col1 && canMoveTo(g.getArea(areaID{thiefArea.Row, thiefArea.Column - 1})) {
		as = append(as, g.getArea(areaID{thiefArea.Row, thiefArea.Column - 1}))
	}

	// Move Right
	if thiefArea.Column+1 <= col8 && canMoveTo(g.getArea(areaID{thiefArea.Row, thiefArea.Column + 1})) {
		as = append(as, g.getArea(areaID{thiefArea.Row, thiefArea.Column + 1}))
	}

	// Move Up
	if thiefArea.Row-1 >= rowA && canMoveTo(g.getArea(areaID{thiefArea.Row - 1, thiefArea.Column})) {
		as = append(as, g.getArea(areaID{thiefArea.Row - 1, thiefArea.Column}))
	}

	// Move Down
	if thiefArea.Row+1 <= g.lastRow() && canMoveTo(g.getArea(areaID{thiefArea.Row + 1, thiefArea.Column})) {
		as = append(as, g.getArea(areaID{thiefArea.Row + 1, thiefArea.Column}))
	}

	return as
}

func (g *game) coinsAreas(thiefArea *Area) []*Area {
	return g.turban1Areas(thiefArea)
}
