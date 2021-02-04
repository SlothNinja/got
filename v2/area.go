package main

import (
	"fmt"
)

type grid [][]*Area

func (g grid) each(f func(*Area)) {
	for row := range g {
		for _, a := range g[row] {
			f(a)
		}
	}
}

type aRow int

const (
	noRow aRow = iota
	rowA
	rowB
	rowC
	rowD
	rowE
	rowF
	rowG
)

var rowIDStrings = [8]string{"None", "A", "B", "C", "D", "E", "F", "G"}

func (r aRow) String() string {
	if r >= rowA && r <= rowG {
		return rowIDStrings[r]
	}
	return "None"
}

type aCol int

const (
	noCol aCol = iota
	col1
	col2
	col3
	col4
	col5
	col6
	col7
	col8
)

func (c aCol) String() string {
	if c >= col1 && c <= col8 {
		return fmt.Sprintf("%d", c)
	}
	return "None"
}

type areaID struct {
	Row    aRow `json:"row"`
	Column aCol `json:"column"`
}

func (id areaID) String() string {
	return fmt.Sprintf("%s-%s", id.Row, id.Column)
}

// Area of the grid.
type Area struct {
	areaID
	Thief     int   `json:"thief"`
	Card      *Card `json:"card"`
	Clickable bool  `json:"clickable"`
}

// selectedThiefArea returns the area corresponding to a previously selected thief.
func (cl *client) selectedThiefArea() *Area {
	return cl.area(cl.g.thiefAreaID)
}

func (cl *client) area(id areaID) *Area {
	if id.Row < rowA || id.Row > cl.lastRow() {
		return nil
	}
	if id.Column < col1 || id.Column > col8 {
		return nil
	}
	return cl.g.grid[id.Row-1][id.Column-1]
}

func newArea(id areaID, card *Card) *Area {
	return &Area{
		areaID: id,
		Thief:  noPID,
		Card:   card,
	}
}

func (cl *client) lastRow() aRow {
	row := rowG
	if cl.g.NumPlayers == 2 {
		row = rowF
	}
	return row
}

func (cl *client) createGrid() {
	deck := newDeck()
	cl.g.grid = make(grid, cl.lastRow())
	for row := rowA; row <= cl.lastRow(); row++ {
		cl.g.grid[row-1] = make([]*Area, 8)
		for col := col1; col <= col8; col++ {
			cl.g.grid[row-1][col-1] = newArea(areaID{row, col}, deck.draw())
		}
	}
}

func (a *Area) hasThief() bool {
	return a.Thief != noPID
}

func (a *Area) hasCard() bool {
	return a.Card != nil
}

func hasArea(as []*Area, a2 *Area) bool {
	for _, a1 := range as {
		if a1.Row == a2.Row && a1.Column == a2.Column {
			return true
		}
	}
	return false
}

func (cl *client) updateClickablesFor(p *player, ta *Area) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	canClick := cl.canClick(p, ta)
	cl.g.grid.each(func(a *Area) { a.Clickable = canClick(a) })
}

// canClick a function specialized by current game context to test whether a player can click on
// a particular area in the grid.  The main benefit is the function provides a closure around area computions,
// essentially caching the results.
func (cl *client) canClick(p *player, ta *Area) func(*Area) bool {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	ff := func(a *Area) bool { return false }
	err := cl.validatePlayerAction()
	switch {
	case cl.g == nil:
		return ff
	case err != nil:
		return ff
	case cl.g.Phase == placeThievesPhase:
		return func(a *Area) bool { return a.Thief == noPID }
	case cl.g.Phase == selectThiefPhase:
		return func(a *Area) bool { return a.Thief == cl.cp.ID }
	case cl.g.Phase == moveThiefPhase:
		switch {
		case p == nil:
			return ff
		case p.ID != cl.cp.ID:
			return ff
		case cl.g.playedCard == nil:
			return ff
		case ta == nil:
			return ff
		case cl.g.playedCard.Kind == lampCard || cl.g.playedCard.Kind == sLampCard:
			as := cl.lampAreas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case cl.g.playedCard.Kind == camelCard || cl.g.playedCard.Kind == sCamelCard:
			as := cl.camelAreas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case cl.g.playedCard.Kind == swordCard:
			as := cl.swordAreas(cl.cp, ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case cl.g.playedCard.Kind == carpetCard:
			as := cl.carpetAreas()
			return func(a *Area) bool { return hasArea(as, a) }
		case cl.g.playedCard.Kind == turbanCard && cl.g.stepped == 0:
			as := cl.turban0Areas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case cl.g.playedCard.Kind == turbanCard && cl.g.stepped == 1:
			as := cl.turban1Areas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case cl.g.playedCard.Kind == coinsCard:
			as := cl.coinsAreas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		default:
			return ff
		}
	default:
		return ff
	}
}
