package main

import (
	"fmt"

	"github.com/SlothNinja/user"
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
func (g *Game) selectedThiefArea() *Area {
	return g.area(g.thiefAreaID)
}

func (g *Game) area(id areaID) *Area {
	if id.Row < rowA || id.Row > g.lastRow() {
		return nil
	}
	if id.Column < col1 || id.Column > col8 {
		return nil
	}
	return g.grid[id.Row-1][id.Column-1]
}

func newArea(id areaID, card *Card) *Area {
	return &Area{
		areaID: id,
		Thief:  noPID,
		Card:   card,
	}
}

func (g *Game) lastRow() aRow {
	row := rowG
	if g.NumPlayers == 2 {
		row = rowF
	}
	return row
}

func (g *Game) createGrid() {
	deck := newDeck()
	g.grid = make(grid, g.lastRow())
	for row := rowA; row <= g.lastRow(); row++ {
		g.grid[row-1] = make([]*Area, 8)
		for col := col1; col <= col8; col++ {
			g.grid[row-1][col-1] = newArea(areaID{row, col}, deck.draw())
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

func (g *Game) updateClickablesFor(cu *user.User, ta *Area) {
	canClick := g.canClick(cu, ta)
	g.grid.each(func(a *Area) { a.Clickable = canClick(a) })
}

// canClick a function specialized by current game context to test whether a player can click on
// a particular area in the grid.  The main benefit is the function provides a closure around area computions,
// essentially caching the results.
func (g *Game) canClick(cu *user.User, ta *Area) func(*Area) bool {
	ff := func(a *Area) bool { return false }
	cp, err := g.validatePlayerAction(cu)
	switch {
	case g == nil:
		return ff
	case err != nil:
		return ff
	case g.Phase == placeThievesPhase:
		return func(a *Area) bool { return a.Thief == noPID }
	case g.Phase == selectThiefPhase:
		return func(a *Area) bool { return a.Thief == cp.ID }
	case g.Phase == moveThiefPhase:
		switch {
		case cp == nil:
			return ff
		case cp.ID != cp.ID:
			return ff
		case g.playedCard == nil:
			return ff
		case ta == nil:
			return ff
		case g.playedCard.Kind == lampCard || g.playedCard.Kind == sLampCard:
			as := g.lampAreas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.playedCard.Kind == camelCard || g.playedCard.Kind == sCamelCard:
			as := g.camelAreas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.playedCard.Kind == swordCard:
			as := g.swordAreas(cp, ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.playedCard.Kind == carpetCard:
			as := g.carpetAreas()
			return func(a *Area) bool { return hasArea(as, a) }
		case g.playedCard.Kind == turbanCard && g.stepped == 0:
			as := g.turban0Areas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.playedCard.Kind == turbanCard && g.stepped == 1:
			as := g.turban1Areas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.playedCard.Kind == coinsCard:
			as := g.coinsAreas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		default:
			return ff
		}
	default:
		return ff
	}
}
