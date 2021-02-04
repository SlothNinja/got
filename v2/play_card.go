package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (cl *client) playCardHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.getGame()
	if err != nil {
		cl.jerr(err)
		return
	}

	err = cl.playCard()
	if err != nil {
		cl.jerr(err)
		return
	}

	ks, es := cl.g.cache()
	_, err = cl.DS.Put(c, ks, es)
	if err != nil {
		cl.jerr(err)
		return
	}

	cl.updateClickablesFor(cl.cp, cl.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": cl.g})
}

func (cl *client) playCard() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	// reset card played related flags
	cl.g.stepped = 0
	cl.g.playedCard = nil

	card, err := cl.validatePlayCard()
	if err != nil {
		return err
	}

	cl.cp.Hand.play(card)
	cl.cp.DiscardPile = append(Cards{card}, cl.cp.DiscardPile...)
	cl.cp.Stats.Played.inc(card.Kind)

	if card.Kind == jewels {
		pc := cl.g.jewels
		cl.cp.Stats.JewelsAs.inc(pc.Kind)
		cl.g.playedCard = &pc
	} else {
		cl.g.playedCard = card
	}

	cl.g.Phase = selectThiefPhase
	cl.g.Undo.Update()

	cl.g.newEntryFor(cl.cp.ID, message{
		"template": "play-card",
		"card":     *card,
	})
	return nil
}

func (cl *client) validatePlayCard() (*Card, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validatePlayerAction()
	if err != nil {
		return nil, err
	}

	card, err := cl.getCard()
	switch {
	case err != nil:
		return nil, err
	case card == nil:
		return nil, fmt.Errorf("you must select a card: %w", sn.ErrValidation)
	case cl.g.Phase != playCardPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w", playCardPhase, cl.g.Phase, sn.ErrValidation)
	default:
		return card, nil
	}
}

func (cl *client) getCard() (*Card, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	obj := struct {
		Kind cKind `json:"kind"`
	}{}

	err := cl.ctx.Bind(&obj)
	if err != nil {
		return nil, err
	}

	i, found := cl.cp.Hand.indexFor(newCard(obj.Kind, false))
	if !found {
		return nil, fmt.Errorf("unable to find card: %w", sn.ErrValidation)
	}
	return cl.cp.Hand[i], nil
}

func (cl *client) lampAreas(thiefArea *Area) []*Area {
	var as []*Area

	// Move Left
	var a2 *Area
	for col := thiefArea.Column - 1; col >= col1; col-- {
		temp := cl.area(areaID{thiefArea.Row, col})
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
		temp := cl.area(areaID{thiefArea.Row, col})
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
		temp := cl.area(areaID{row, thiefArea.Column})
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
	for row := thiefArea.Row + 1; row <= cl.lastRow(); row++ {
		temp := cl.area(areaID{row, thiefArea.Column})
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
func (cl *client) camelAreas(ta *Area) []*Area {
	var as []*Area

	// Move Three Left?
	if ta.Column-3 >= col1 {
		area1 := cl.area(areaID{ta.Row, ta.Column - 1})
		area2 := cl.area(areaID{ta.Row, ta.Column - 2})
		area3 := cl.area(areaID{ta.Row, ta.Column - 3})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Three Right?
	if ta.Column+3 <= col8 {
		area1 := cl.area(areaID{ta.Row, ta.Column + 1})
		area2 := cl.area(areaID{ta.Row, ta.Column + 2})
		area3 := cl.area(areaID{ta.Row, ta.Column + 3})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Three Up?
	if ta.Row-3 >= rowA {
		area1 := cl.area(areaID{ta.Row - 1, ta.Column})
		area2 := cl.area(areaID{ta.Row - 2, ta.Column})
		area3 := cl.area(areaID{ta.Row - 3, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Three Down?
	if ta.Row+3 <= cl.lastRow() {
		area1 := cl.area(areaID{ta.Row + 1, ta.Column})
		area2 := cl.area(areaID{ta.Row + 2, ta.Column})
		area3 := cl.area(areaID{ta.Row + 3, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Left One Up or One Up Two Left or One Left One Up One Left?
	if ta.Column-2 >= col1 && ta.Row-1 >= rowA {
		area1 := cl.area(areaID{ta.Row, ta.Column - 1})
		area2 := cl.area(areaID{ta.Row, ta.Column - 2})
		area3 := cl.area(areaID{ta.Row - 1, ta.Column - 2})
		area4 := cl.area(areaID{ta.Row - 1, ta.Column})
		area5 := cl.area(areaID{ta.Row - 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Left One Down or One Down Two Left or One Left One Down One Left?
	if ta.Column-2 >= col1 && ta.Row+1 <= cl.lastRow() {
		area1 := cl.area(areaID{ta.Row, ta.Column - 1})
		area2 := cl.area(areaID{ta.Row, ta.Column - 2})
		area3 := cl.area(areaID{ta.Row + 1, ta.Column - 2})
		area4 := cl.area(areaID{ta.Row + 1, ta.Column})
		area5 := cl.area(areaID{ta.Row + 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Right One Up or One Up Two Right or One Right One Up One Right?
	if ta.Column+2 <= col8 && ta.Row-1 >= rowA {
		area1 := cl.area(areaID{ta.Row, ta.Column + 1})
		area2 := cl.area(areaID{ta.Row, ta.Column + 2})
		area3 := cl.area(areaID{ta.Row - 1, ta.Column + 2})
		area4 := cl.area(areaID{ta.Row - 1, ta.Column})
		area5 := cl.area(areaID{ta.Row - 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Right One Down or One Down Two Right or One Right One Down One Right?
	if ta.Column+2 <= col8 && ta.Row+1 <= cl.lastRow() {
		area1 := cl.area(areaID{ta.Row, ta.Column + 1})
		area2 := cl.area(areaID{ta.Row, ta.Column + 2})
		area3 := cl.area(areaID{ta.Row + 1, ta.Column + 2})
		area4 := cl.area(areaID{ta.Row + 1, ta.Column})
		area5 := cl.area(areaID{ta.Row + 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Right Two Down or Two Down One Right or One Down One Right One Down?
	if ta.Column+1 <= col8 && ta.Row+2 <= cl.lastRow() {
		area1 := cl.area(areaID{ta.Row + 1, ta.Column})
		area2 := cl.area(areaID{ta.Row + 2, ta.Column})
		area3 := cl.area(areaID{ta.Row + 2, ta.Column + 1})
		area4 := cl.area(areaID{ta.Row, ta.Column + 1})
		area5 := cl.area(areaID{ta.Row + 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Right Two Up or Two Up One Right or One Up One Right One Up?
	if ta.Column+1 <= col8 && ta.Row-2 >= rowA {
		area1 := cl.area(areaID{ta.Row - 1, ta.Column})
		area2 := cl.area(areaID{ta.Row - 2, ta.Column})
		area3 := cl.area(areaID{ta.Row - 2, ta.Column + 1})
		area4 := cl.area(areaID{ta.Row, ta.Column + 1})
		area5 := cl.area(areaID{ta.Row - 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Left Two Down or Two Down One Left or One Down One Left One Down?
	if ta.Column-1 >= col1 && ta.Row+2 <= cl.lastRow() {
		area1 := cl.area(areaID{ta.Row + 1, ta.Column})
		area2 := cl.area(areaID{ta.Row + 2, ta.Column})
		area3 := cl.area(areaID{ta.Row + 2, ta.Column - 1})
		area4 := cl.area(areaID{ta.Row, ta.Column - 1})
		area5 := cl.area(areaID{ta.Row + 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Left Two Up or Two Up One Left or One Up One Left One Up?
	if ta.Column-1 >= col1 && ta.Row-2 >= rowA {
		area1 := cl.area(areaID{ta.Row - 1, ta.Column})
		area2 := cl.area(areaID{ta.Row - 2, ta.Column})
		area3 := cl.area(areaID{ta.Row - 2, ta.Column - 1})
		area4 := cl.area(areaID{ta.Row, ta.Column - 1})
		area5 := cl.area(areaID{ta.Row - 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Left One Up One Right or One Up One Left One Down?
	if ta.Column-1 >= col1 && ta.Row-1 >= rowA {
		area1 := cl.area(areaID{ta.Row, ta.Column - 1})
		area2 := cl.area(areaID{ta.Row - 1, ta.Column - 1})
		area3 := cl.area(areaID{ta.Row - 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	// Move One Up One Right One Down or One Right One Up One Left?
	if ta.Column+1 <= col8 && ta.Row-1 >= rowA {
		area1 := cl.area(areaID{ta.Row, ta.Column + 1})
		area2 := cl.area(areaID{ta.Row - 1, ta.Column + 1})
		area3 := cl.area(areaID{ta.Row - 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	// Move One Left One Down One Right or One Down One Left One Up?
	if ta.Column-1 >= col1 && ta.Row+1 <= cl.lastRow() {
		area1 := cl.area(areaID{ta.Row, ta.Column - 1})
		area2 := cl.area(areaID{ta.Row + 1, ta.Column - 1})
		area3 := cl.area(areaID{ta.Row + 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	// Move One Down One Right One Up or One Right One Down One Left?
	if ta.Column+1 <= col8 && ta.Row+1 <= cl.lastRow() {
		area1 := cl.area(areaID{ta.Row, ta.Column + 1})
		area2 := cl.area(areaID{ta.Row + 1, ta.Column + 1})
		area3 := cl.area(areaID{ta.Row + 1, ta.Column})
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

func (cl *client) swordAreas(cp *player, thiefArea *Area) []*Area {
	var as []*Area

	// Move Left
	if area, row := thiefArea, thiefArea.Row; thiefArea.Column >= col3 {
		// Left as far as permitted
		for col := thiefArea.Column - 1; col >= col3; col-- {
			if temp := cl.area(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := cl.area(areaID{row, area.Column - 1}), cl.area(areaID{row, area.Column - 2})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	// Move Right
	if area, row := thiefArea, thiefArea.Row; thiefArea.Column <= col6 {
		// Right as far as permitted
		for col := thiefArea.Column + 1; col <= col6; col++ {
			if temp := cl.area(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := cl.area(areaID{row, area.Column + 1}), cl.area(areaID{row, area.Column + 2})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	// Move Up
	if area, col := thiefArea, thiefArea.Column; thiefArea.Row >= rowC {
		// Up as far as permitted
		for row := thiefArea.Row - 1; row >= rowC; row-- {
			if temp := cl.area(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := cl.area(areaID{area.Row - 1, col}), cl.area(areaID{area.Row - 2, col})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	// Move Down
	if area, col := thiefArea, thiefArea.Column; thiefArea.Row <= cl.lastRow()-2 {
		// Down as far as permitted
		for row := thiefArea.Row + 1; row <= cl.lastRow()-2; row++ {
			//g.debugf("Row: %v Col: %v", row, col)
			if temp := cl.area(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := cl.area(areaID{area.Row + 1, col}), cl.area(areaID{area.Row + 2, col})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	return as
}

func (p *player) anotherThiefIn(a *Area) bool {
	return a.hasThief() && a.Thief != p.ID
}

func (cl *client) isCarpetArea(a *Area) bool {
	if cl.selectedThiefArea() != nil {
		return hasArea(cl.carpetAreas(), a)
	}
	return false
}

func (cl *client) carpetAreas() []*Area {
	as := make([]*Area, 0)
	a1 := cl.selectedThiefArea()

	// Move Left
	var a2, empty *Area
MoveLeft:
	for col := a1.Column - 1; col >= col1; col-- {
		switch temp := cl.area(areaID{a1.Row, col}); {
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
		switch temp := cl.area(areaID{a1.Row, col}); {
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
		switch temp := cl.area(areaID{row, a1.Column}); {
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
	for row := a1.Row + 1; row <= cl.lastRow(); row++ {
		switch temp := cl.area(areaID{row, a1.Column}); {
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

func (cl *client) turban0Areas(thiefArea *Area) []*Area {
	var as []*Area

	// Move Left
	if col := thiefArea.Column - 1; col >= col1 {
		if area := cl.area(areaID{thiefArea.Row, col}); canMoveTo(area) {
			// Left
			if col := col - 1; col >= col1 && canMoveTo(cl.area(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Up
			if row := area.Row - 1; row >= rowA && canMoveTo(cl.area(areaID{row, col})) {
				as = append(as, area)
			}
			// Down
			if row := area.Row + 1; row <= cl.lastRow() && canMoveTo(cl.area(areaID{row, col})) {
				as = append(as, area)
			}
		}
	}

	// Move Right
	if col := thiefArea.Column + 1; col <= col8 {
		if area := cl.area(areaID{thiefArea.Row, col}); canMoveTo(area) {
			// Right
			if col := col + 1; col <= col8 && canMoveTo(cl.area(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Up
			if row := area.Row - 1; row >= rowA && canMoveTo(cl.area(areaID{row, col})) {
				as = append(as, area)
			}
			// Down
			if row := area.Row + 1; row <= cl.lastRow() && canMoveTo(cl.area(areaID{row, col})) {
				as = append(as, area)
			}
		}
	}

	// Move Up
	if row := thiefArea.Row - 1; row >= rowA {
		if area := cl.area(areaID{row, thiefArea.Column}); canMoveTo(area) {
			// Left
			if col := area.Column - 1; col >= col1 && canMoveTo(cl.area(areaID{row, col})) {
				as = append(as, area)
			}
			// Right
			if col := area.Column + 1; col <= col8 && canMoveTo(cl.area(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Up
			if row := row - 1; row >= rowA && canMoveTo(cl.area(areaID{row, thiefArea.Column})) {
				as = append(as, area)
			}
		}
	}

	// Move Down
	if row := thiefArea.Row + 1; row <= cl.lastRow() {
		if area := cl.area(areaID{row, thiefArea.Column}); canMoveTo(area) {
			// Left
			if col := area.Column - 1; col >= col1 && canMoveTo(cl.area(areaID{row, col})) {
				as = append(as, area)
			}
			// Right
			if col := area.Column + 1; col <= col8 && canMoveTo(cl.area(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Down
			if row := row + 1; row <= cl.lastRow() && canMoveTo(cl.area(areaID{row, thiefArea.Column})) {
				as = append(as, area)
			}
		}
	}

	return as
}

func (cl *client) turban1Areas(thiefArea *Area) []*Area {
	var as []*Area

	// Move Left
	if thiefArea.Column-1 >= col1 && canMoveTo(cl.area(areaID{thiefArea.Row, thiefArea.Column - 1})) {
		as = append(as, cl.area(areaID{thiefArea.Row, thiefArea.Column - 1}))
	}

	// Move Right
	if thiefArea.Column+1 <= col8 && canMoveTo(cl.area(areaID{thiefArea.Row, thiefArea.Column + 1})) {
		as = append(as, cl.area(areaID{thiefArea.Row, thiefArea.Column + 1}))
	}

	// Move Up
	if thiefArea.Row-1 >= rowA && canMoveTo(cl.area(areaID{thiefArea.Row - 1, thiefArea.Column})) {
		as = append(as, cl.area(areaID{thiefArea.Row - 1, thiefArea.Column}))
	}

	// Move Down
	if thiefArea.Row+1 <= cl.lastRow() && canMoveTo(cl.area(areaID{thiefArea.Row + 1, thiefArea.Column})) {
		as = append(as, cl.area(areaID{thiefArea.Row + 1, thiefArea.Column}))
	}

	return as
}

func (cl *client) coinsAreas(thiefArea *Area) []*Area {
	return cl.turban1Areas(thiefArea)
}
