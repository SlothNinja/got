package main

import (
	"fmt"
)

type Grid [][]*Area

func (g Grid) Each(f func(*Area)) {
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

// SelectedThiefArea returns the area corresponding to a previously selected thief.
func (g *Game) SelectedThiefArea() *Area {
	return g.getArea(g.SelectedThiefAreaID)
}

func (g *Game) getArea(id areaID) *Area {
	if id.Row < rowA || id.Row > g.lastRow() {
		return nil
	}
	if id.Column < col1 || id.Column > col8 {
		return nil
	}
	return g.Grid[id.Row-1][id.Column-1]
}

func newArea(id areaID, card *Card) *Area {
	return &Area{
		areaID: id,
		Thief:  noPID,
		Card:   card,
	}
}

func (h *Game) lastRow() aRow {
	row := rowG
	if h.NumPlayers == 2 {
		row = rowF
	}
	return row
}

func (h *Game) createGrid() {
	deck := newDeck()
	h.Grid = make(Grid, h.lastRow())
	for row := rowA; row <= h.lastRow(); row++ {
		h.Grid[row-1] = make([]*Area, 8)
		for col := col1; col <= col8; col++ {
			h.Grid[row-1][col-1] = newArea(areaID{row, col}, deck.draw())
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
