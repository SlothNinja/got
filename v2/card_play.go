package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

// func init() {
// 	gob.Register(new(playCardEntry))
// }

func (h *History) startCardPlay(c *gin.Context) (tmpl string, err error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	h.Phase = playCard
	h.Turn = 1
	return
}

func (client Client) playCard(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	h, err := client.getHistory(c)
	if err != nil {
		jerr(c, err)
		return
	}

	log.Debugf("h.Key: %v", h.Key)

	err = h.playCard(c)
	if err != nil {
		jerr(c, err)
		return
	}

	ks, es := h.cache()
	_, err = client.DS.Put(c, ks, es)
	if err != nil {
		jerr(c, err)
		return
	}

	h.updateClickablesFor(c, h.CurrentPlayer())
	c.JSON(http.StatusOK, gin.H{"game": h})
}

func (h *History) playCard(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	// reset card played related flags
	h.Stepped = 0
	h.JewelsPlayed = false
	h.PlayedCard = nil

	card, err := h.validatePlayCard(c)
	if err != nil {
		return err
	}

	cp := h.CurrentPlayer()
	cp.Hand.play(card)
	cp.DiscardPile = append(Cards{card}, cp.DiscardPile...)

	if card.Type == jewels {
		pc := h.Jewels
		h.PlayedCard = &pc
		h.JewelsPlayed = true
	} else {
		h.PlayedCard = card
	}

	// Log placement
	// e := g.newPlayCardEntryFor(cp, card)
	// restful.AddNoticef(c, string(e.HTML(g)))

	h.startSelectThief(c)
	h.Undo.Update()
	return nil
}

func (h *History) validatePlayCard(c *gin.Context) (*Card, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := h.validatePlayerAction(c)
	if err != nil {
		return nil, err
	}

	card, err := h.getCardFrom(c)
	switch {
	case err != nil:
		return nil, err
	case card == nil:
		return nil, fmt.Errorf("you must select a card: %w", sn.ErrValidation)
	default:
		return card, nil
	}
}

func (h *History) getCardFrom(c *gin.Context) (*Card, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	obj := struct {
		Kind cKind `json:"kind"`
	}{}

	err := c.Bind(&obj)
	if err != nil {
		return nil, err
	}

	log.Debugf("obj: %#v", obj)

	cp := h.CurrentPlayer()
	i, found := cp.Hand.indexFor(newCard(obj.Kind, false))
	if !found {
		return nil, fmt.Errorf("unable to find card: %#w", sn.ErrValidation)
	}
	return cp.Hand[i], nil
}

// type playCardEntry struct {
// 	*Entry
// 	Type cType
// }

// func (g *History) newPlayCardEntryFor(p *Player, c *Card) *playCardEntry {
// 	e := &playCardEntry{
// 		Entry: g.newEntryFor(p),
// 		Type:  c.Type,
// 	}
// 	p.Log = append(p.Log, e)
// 	g.Log = append(g.Log, e)
// 	return e
// }

// func (e *playCardEntry) HTML(g *History) template.HTML {
// 	return restful.HTML("%s played %s card.", g.NameByPID(e.PlayerID), e.Type)
// }

func (h *History) lampAreas(thiefArea *Area) []*Area {
	var as []*Area

	// Move Left
	var a2 *Area
	for col := thiefArea.Column - 1; col >= col1; col-- {
		temp := h.getArea(areaID{thiefArea.Row, col})
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
		temp := h.getArea(areaID{thiefArea.Row, col})
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
		temp := h.getArea(areaID{row, thiefArea.Column})
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
	for row := thiefArea.Row + 1; row <= h.lastRow(); row++ {
		temp := h.getArea(areaID{row, thiefArea.Column})
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
func (h *History) camelAreas(ta *Area) []*Area {
	var as []*Area

	// Move Three Left?
	if ta.Column-3 >= col1 {
		area1 := h.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := h.getArea(areaID{ta.Row, ta.Column - 2})
		area3 := h.getArea(areaID{ta.Row, ta.Column - 3})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Three Right?
	if ta.Column+3 <= col8 {
		area1 := h.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := h.getArea(areaID{ta.Row, ta.Column + 2})
		area3 := h.getArea(areaID{ta.Row, ta.Column + 3})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Three Up?
	if ta.Row-3 >= rowA {
		area1 := h.getArea(areaID{ta.Row - 1, ta.Column})
		area2 := h.getArea(areaID{ta.Row - 2, ta.Column})
		area3 := h.getArea(areaID{ta.Row - 3, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Three Down?
	if ta.Row+3 <= h.lastRow() {
		area1 := h.getArea(areaID{ta.Row + 1, ta.Column})
		area2 := h.getArea(areaID{ta.Row + 2, ta.Column})
		area3 := h.getArea(areaID{ta.Row + 3, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Left One Up or One Up Two Left or One Left One Up One Left?
	if ta.Column-2 >= col1 && ta.Row-1 >= rowA {
		area1 := h.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := h.getArea(areaID{ta.Row, ta.Column - 2})
		area3 := h.getArea(areaID{ta.Row - 1, ta.Column - 2})
		area4 := h.getArea(areaID{ta.Row - 1, ta.Column})
		area5 := h.getArea(areaID{ta.Row - 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Left One Down or One Down Two Left or One Left One Down One Left?
	if ta.Column-2 >= col1 && ta.Row+1 <= h.lastRow() {
		area1 := h.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := h.getArea(areaID{ta.Row, ta.Column - 2})
		area3 := h.getArea(areaID{ta.Row + 1, ta.Column - 2})
		area4 := h.getArea(areaID{ta.Row + 1, ta.Column})
		area5 := h.getArea(areaID{ta.Row + 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Right One Up or One Up Two Right or One Right One Up One Right?
	if ta.Column+2 <= col8 && ta.Row-1 >= rowA {
		area1 := h.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := h.getArea(areaID{ta.Row, ta.Column + 2})
		area3 := h.getArea(areaID{ta.Row - 1, ta.Column + 2})
		area4 := h.getArea(areaID{ta.Row - 1, ta.Column})
		area5 := h.getArea(areaID{ta.Row - 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move Two Right One Down or One Down Two Right or One Right One Down One Right?
	if ta.Column+2 <= col8 && ta.Row+1 <= h.lastRow() {
		area1 := h.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := h.getArea(areaID{ta.Row, ta.Column + 2})
		area3 := h.getArea(areaID{ta.Row + 1, ta.Column + 2})
		area4 := h.getArea(areaID{ta.Row + 1, ta.Column})
		area5 := h.getArea(areaID{ta.Row + 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Right Two Down or Two Down One Right or One Down One Right One Down?
	if ta.Column+1 <= col8 && ta.Row+2 <= h.lastRow() {
		area1 := h.getArea(areaID{ta.Row + 1, ta.Column})
		area2 := h.getArea(areaID{ta.Row + 2, ta.Column})
		area3 := h.getArea(areaID{ta.Row + 2, ta.Column + 1})
		area4 := h.getArea(areaID{ta.Row, ta.Column + 1})
		area5 := h.getArea(areaID{ta.Row + 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Right Two Up or Two Up One Right or One Up One Right One Up?
	if ta.Column+1 <= col8 && ta.Row-2 >= rowA {
		area1 := h.getArea(areaID{ta.Row - 1, ta.Column})
		area2 := h.getArea(areaID{ta.Row - 2, ta.Column})
		area3 := h.getArea(areaID{ta.Row - 2, ta.Column + 1})
		area4 := h.getArea(areaID{ta.Row, ta.Column + 1})
		area5 := h.getArea(areaID{ta.Row - 1, ta.Column + 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Left Two Down or Two Down One Left or One Down One Left One Down?
	if ta.Column-1 >= col1 && ta.Row+2 <= h.lastRow() {
		area1 := h.getArea(areaID{ta.Row + 1, ta.Column})
		area2 := h.getArea(areaID{ta.Row + 2, ta.Column})
		area3 := h.getArea(areaID{ta.Row + 2, ta.Column - 1})
		area4 := h.getArea(areaID{ta.Row, ta.Column - 1})
		area5 := h.getArea(areaID{ta.Row + 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Left Two Up or Two Up One Left or One Up One Left One Up?
	if ta.Column-1 >= col1 && ta.Row-2 >= rowA {
		area1 := h.getArea(areaID{ta.Row - 1, ta.Column})
		area2 := h.getArea(areaID{ta.Row - 2, ta.Column})
		area3 := h.getArea(areaID{ta.Row - 2, ta.Column - 1})
		area4 := h.getArea(areaID{ta.Row, ta.Column - 1})
		area5 := h.getArea(areaID{ta.Row - 1, ta.Column - 1})
		if canMoveTo(area1, area2, area3) || canMoveTo(area3, area4, area5) || canMoveTo(area1, area5, area3) {
			as = append(as, area3)
		}
	}

	// Move One Left One Up One Right or One Up One Left One Down?
	if ta.Column-1 >= col1 && ta.Row-1 >= rowA {
		area1 := h.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := h.getArea(areaID{ta.Row - 1, ta.Column - 1})
		area3 := h.getArea(areaID{ta.Row - 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	// Move One Up One Right One Down or One Right One Up One Left?
	if ta.Column+1 <= col8 && ta.Row-1 >= rowA {
		area1 := h.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := h.getArea(areaID{ta.Row - 1, ta.Column + 1})
		area3 := h.getArea(areaID{ta.Row - 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	// Move One Left One Down One Right or One Down One Left One Up?
	if ta.Column-1 >= col1 && ta.Row+1 <= h.lastRow() {
		area1 := h.getArea(areaID{ta.Row, ta.Column - 1})
		area2 := h.getArea(areaID{ta.Row + 1, ta.Column - 1})
		area3 := h.getArea(areaID{ta.Row + 1, ta.Column})
		if canMoveTo(area1, area2, area3) {
			as = append(as, area1)
			as = append(as, area3)
		}
	}

	// Move One Down One Right One Up or One Right One Down One Left?
	if ta.Column+1 <= col8 && ta.Row+1 <= h.lastRow() {
		area1 := h.getArea(areaID{ta.Row, ta.Column + 1})
		area2 := h.getArea(areaID{ta.Row + 1, ta.Column + 1})
		area3 := h.getArea(areaID{ta.Row + 1, ta.Column})
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

func (h *History) isSwordArea(a *Area) bool {
	if h.SelectedThiefArea() != nil {
		return hasArea(h.swordAreas(), a)
	}
	return false
}

func (h *History) swordAreas() []*Area {
	cp := h.CurrentPlayer()
	if h.ClickAreas != nil {
		return h.ClickAreas
	}
	as := make([]*Area, 0)
	a := h.SelectedThiefArea()

	// Move Left
	if area, row := a, a.Row; a.Column >= col3 {
		// Left as far as permitted
		for col := a.Column - 1; col >= col3; col-- {
			if temp := h.getArea(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := h.getArea(areaID{row, area.Column - 1}), h.getArea(areaID{row, area.Column - 2})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	// Move Right
	if area, row := a, a.Row; a.Column <= col6 {
		// Right as far as permitted
		for col := a.Column + 1; col <= col6; col++ {
			if temp := h.getArea(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := h.getArea(areaID{row, area.Column + 1}), h.getArea(areaID{row, area.Column + 2})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	// Move Up
	if area, col := a, a.Column; a.Row >= rowC {
		// Up as far as permitted
		for row := a.Row - 1; row >= rowC; row-- {
			if temp := h.getArea(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := h.getArea(areaID{area.Row - 1, col}), h.getArea(areaID{area.Row - 2, col})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	// Move Down
	if area, col := a, a.Column; a.Row <= h.lastRow()-2 {
		// Down as far as permitted
		for row := a.Row + 1; row <= h.lastRow()-2; row++ {
			//g.debugf("Row: %v Col: %v", row, col)
			if temp := h.getArea(areaID{row, col}); !canMoveTo(temp) {
				break
			} else {
				area = temp
			}
		}

		// Check for Thief and Place to Bump
		moveTo, bumpTo := h.getArea(areaID{area.Row + 1, col}), h.getArea(areaID{area.Row + 2, col})
		if cp.anotherThiefIn(moveTo) && canMoveTo(bumpTo) {
			as = append(as, moveTo)
		}
	}

	h.ClickAreas = as
	return as
}

func (p *Player) anotherThiefIn(a *Area) bool {
	return a.hasThief() && a.Thief != p.ID
}

func (h *History) isCarpetArea(a *Area) bool {
	if h.SelectedThiefArea() != nil {
		return hasArea(h.carpetAreas(), a)
	}
	return false
}

func (h *History) carpetAreas() []*Area {
	if h.ClickAreas != nil {
		return h.ClickAreas
	}
	as := make([]*Area, 0)
	a1 := h.SelectedThiefArea()

	// Move Left
	var a2, empty *Area
MoveLeft:
	for col := a1.Column - 1; col >= col1; col-- {
		switch temp := h.getArea(areaID{a1.Row, col}); {
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
		switch temp := h.getArea(areaID{a1.Row, col}); {
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
		switch temp := h.getArea(areaID{row, a1.Column}); {
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
	for row := a1.Row + 1; row <= h.lastRow(); row++ {
		switch temp := h.getArea(areaID{row, a1.Column}); {
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

	h.ClickAreas = as
	return as
}

func (h *History) isTurban0Area(a *Area) bool {
	if h.SelectedThiefArea() != nil {
		return hasArea(h.turban0Areas(), a)
	}
	return false
}

func (h *History) turban0Areas() []*Area {
	if h.ClickAreas != nil {
		return h.ClickAreas
	}
	as := make([]*Area, 0)
	a := h.SelectedThiefArea()

	// Move Left
	if col := a.Column - 1; col >= col1 {
		if area := h.getArea(areaID{a.Row, col}); canMoveTo(area) {
			// Left
			if col := col - 1; col >= col1 && canMoveTo(h.getArea(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Up
			if row := area.Row - 1; row >= rowA && canMoveTo(h.getArea(areaID{row, col})) {
				as = append(as, area)
			}
			// Down
			if row := area.Row + 1; row <= h.lastRow() && canMoveTo(h.getArea(areaID{row, col})) {
				as = append(as, area)
			}
		}
	}

	// Move Right
	if col := a.Column + 1; col <= col8 {
		if area := h.getArea(areaID{a.Row, col}); canMoveTo(area) {
			// Right
			if col := col + 1; col <= col8 && canMoveTo(h.getArea(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Up
			if row := area.Row - 1; row >= rowA && canMoveTo(h.getArea(areaID{row, col})) {
				as = append(as, area)
			}
			// Down
			if row := area.Row + 1; row <= h.lastRow() && canMoveTo(h.getArea(areaID{row, col})) {
				as = append(as, area)
			}
		}
	}

	// Move Up
	if row := a.Row - 1; row >= rowA {
		if area := h.getArea(areaID{row, a.Column}); canMoveTo(area) {
			// Left
			if col := area.Column - 1; col >= col1 && canMoveTo(h.getArea(areaID{row, col})) {
				as = append(as, area)
			}
			// Right
			if col := area.Column + 1; col <= col8 && canMoveTo(h.getArea(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Up
			if row := row - 1; row >= rowA && canMoveTo(h.getArea(areaID{row, a.Column})) {
				as = append(as, area)
			}
		}
	}

	// Move Down
	if row := a.Row + 1; row <= h.lastRow() {
		if area := h.getArea(areaID{row, a.Column}); canMoveTo(area) {
			// Left
			if col := area.Column - 1; col >= col1 && canMoveTo(h.getArea(areaID{row, col})) {
				as = append(as, area)
			}
			// Right
			if col := area.Column + 1; col <= col8 && canMoveTo(h.getArea(areaID{area.Row, col})) {
				as = append(as, area)
			}
			// Down
			if row := row + 1; row <= h.lastRow() && canMoveTo(h.getArea(areaID{row, a.Column})) {
				as = append(as, area)
			}
		}
	}

	h.ClickAreas = as
	return as
}

func (h *History) isTurban1Area(a *Area) bool {
	if h.SelectedThiefArea() != nil {
		return hasArea(h.turban1Areas(), a)
	}
	return false
}

func (h *History) turban1Areas() []*Area {
	if h.ClickAreas != nil {
		return h.ClickAreas
	}
	as := make([]*Area, 0)
	a := h.SelectedThiefArea()

	// Move Left
	if a.Column-1 >= col1 && canMoveTo(h.getArea(areaID{a.Row, a.Column - 1})) {
		as = append(as, h.getArea(areaID{a.Row, a.Column - 1}))
	}

	// Move Right
	if a.Column+1 <= col8 && canMoveTo(h.getArea(areaID{a.Row, a.Column + 1})) {
		as = append(as, h.getArea(areaID{a.Row, a.Column + 1}))
	}

	// Move Up
	if a.Row-1 >= rowA && canMoveTo(h.getArea(areaID{a.Row - 1, a.Column})) {
		as = append(as, h.getArea(areaID{a.Row - 1, a.Column}))
	}

	// Move Down
	if a.Row+1 <= h.lastRow() && canMoveTo(h.getArea(areaID{a.Row + 1, a.Column})) {
		as = append(as, h.getArea(areaID{a.Row + 1, a.Column}))
	}

	h.ClickAreas = as
	return as
}

func (h *History) isCoinsArea(a *Area) bool {
	if h.SelectedThiefArea() != nil {
		return hasArea(h.coinsAreas(), a)
	}
	return false
}

func (h *History) coinsAreas() []*Area {
	return h.turban1Areas()
}
