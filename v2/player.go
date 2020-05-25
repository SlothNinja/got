package main

import (
	"sort"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

// Player represents one of the players of the game.
type Player struct {
	ID              int        `json:"id"`
	PerformedAction bool       `json:"performedAction"`
	Score           int        `json:"score"`
	Passed          bool       `json:"passed"`
	Colors          []sn.Color `json:"colors"`
	User            *sn.User   `json:"user"`
	Hand            Cards      `json:"hand"`
	DrawPile        Cards      `json:"drawPile"`
	DiscardPile     Cards      `json:"discardPile"`
}

func (g *Game) pids() []int {
	pids := make([]int, len(g.Players))
	for i := range g.Players {
		pids[i] = g.Players[i].ID
	}
	return pids
}

// Players is a slice of players of the game.
type Players []*Player

func allPassed(ps []*Player) bool {
	for _, p := range ps {
		if !p.Passed {
			return false
		}
	}
	return true
}

// Len is part of the sort.Interface interface
func (ps Players) Len() int { return len(ps) }

// Swap is part of the sort.Interface interface
func (ps Players) Swap(i, j int) { ps[i], ps[j] = ps[j], ps[i] }

// ByScore enables sorting players by their score.
type ByScore struct{ Players }

// Less defines when a player has a lower score than another player.
func (bs ByScore) Less(i, j int) bool {
	return bs.Players[i].compareByScore(bs.Players[j]) == sn.LessThan
}

func (p *Player) CompareByScore(p2 *Player) sn.Comparison {
	switch {
	case p.Score < p2.Score:
		return sn.LessThan
	case p.Score > p2.Score:
		return sn.GreaterThan
	default:
		return sn.EqualTo
	}
}

func (p *Player) compareByScore(p2 *Player) sn.Comparison {
	byScore := p.CompareByScore(p2)
	if byScore != sn.EqualTo {
		return byScore
	}

	byLamps := p.compareByLamps(p2)
	if byLamps != sn.EqualTo {
		return byLamps
	}
	byCamels := p.compareByCamels(p2)
	if byCamels != sn.EqualTo {
		return byCamels
	}
	return p.compareByCards(p2)
}

func (p *Player) compareByLamps(p2 *Player) sn.Comparison {
	switch c0, c1 := lampCount(p.Hand...), lampCount(p2.Hand...); {
	case c0 < c1:
		return sn.LessThan
	case c0 > c1:
		return sn.GreaterThan
	}
	return sn.EqualTo
}

// CountFor provides the number of faceUp and faceDown cards a player has.
func (cs Cards) CountFor(t cKind) (faceUp, faceDown int) {
	for _, c := range cs {
		switch {
		case c.Kind == t && c.FaceUp:
			faceUp++
		case c.Kind == t && !c.FaceUp:
			faceDown++
		}
	}
	return
}

func lampCount(cs ...*Card) (count int) {
	for _, c := range cs {
		if c.Kind == lampCard || c.Kind == sLampCard {
			count++
		}
	}
	return count
}

func (p *Player) compareByCamels(p2 *Player) sn.Comparison {
	switch c0, c1 := camelCount(p.Hand...), camelCount(p2.Hand...); {
	case c0 < c1:
		return sn.LessThan
	case c0 > c1:
		return sn.GreaterThan
	}
	return sn.EqualTo
}

func camelCount(cs ...*Card) (count int) {
	for _, c := range cs {
		if c.Kind == camelCard || c.Kind == sCamelCard {
			count++
		}
	}
	return count
}

func (p *Player) compareByCards(p2 *Player) sn.Comparison {
	switch c0, c1 := len(p.Hand), len(p2.Hand); {
	case c0 < c1:
		return sn.LessThan
	case c0 > c1:
		return sn.GreaterThan
	}
	return sn.EqualTo
}

func (client Client) determinePlaces(c *gin.Context, g *Game) (sn.Places, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	// sort players by score
	sort.Sort(Reverse{ByScore{g.Players}})

	places := make(sn.Places, 0)
	rmap := make(sn.ResultsMap, 0)
	for i, p1 := range g.Players {
		results := make(sn.Results, 0)
		tie := false
		for j, p2 := range g.Players {
			r, err := client.Game.For(c, p2.User.Key, g.Type)
			if err != nil {
				return nil, err
			}
			result := &sn.Result{
				GameID: g.ID(),
				Type:   g.Type,
				R:      r.R,
				RD:     r.RD,
			}
			switch c := p1.compareByScore(p2); {
			case i == j:
			case c == sn.GreaterThan:
				result.Outcome = 1
				results = append(results, result)
			case c == sn.LessThan:
				result.Outcome = 0
				results = append(results, result)
			case c == sn.EqualTo:
				result.Outcome = 0.5
				results = append(results, result)
				tie = true
			}
		}
		rmap[p1.User.Key] = results
		if !tie {
			places = append(places, rmap)
			rmap = make(sn.ResultsMap, 0)
		} else if i == len(g.Players)-1 {
			places = append(places, rmap)
		}
	}
	return places, nil
}

// Reverse is a wrapper for sorting in reverse order.
type Reverse struct{ sort.Interface }

// Less indicates if one element should preceed another.
func (r Reverse) Less(i, j int) bool { return r.Interface.Less(j, i) }

// func (p *Player) init(gr sn.Gamer) {
// 	p.SetGame(gr)
// }

func (g *Game) newPlayer(i int) *Player {
	return &Player{
		ID:          i + 1,
		Colors:      defaultColors()[:g.NumPlayers],
		Hand:        newStartHand(),
		DrawPile:    make(Cards, 0),
		DiscardPile: make(Cards, 0),
		User:        sn.ToUser(g.UserKeys[i], g.UserNames[i], g.UserEmailHashes[i]),
	}
}

// func (g *Game) addNewPlayer(i int) {
// 	p := newPlayer()
// 	g.Players = append(g.Players, p)
// 	p.ID = i + 1
// 	p.User = sn.ToUser(g.UserKeys[i], g.UserNames[i], g.UserEmails[i])
// }

// func createPlayer(g *Game, uid int64) *Player {
// 	p := newPlayer()
// 	p.ID = len(g.Players)
// 	return p
// }

func (p *Player) beginningOfTurnReset() {
	p.PerformedAction = false
}

func (h *Game) updateClickablesFor(c *gin.Context, p *Player, ta *Area) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	canClick := h.CanClick(c, p, ta)
	h.Grid.Each(func(a *Area) { a.Clickable = canClick(a) })
}

// CanClick a function specialized by current game context to test whether a player can click on
// a particular area in the grid.  The main benefit is the function provides a closure around area computions,
// essentially caching the results.
func (g *Game) CanClick(c *gin.Context, p *Player, ta *Area) func(*Area) bool {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	ff := func(a *Area) bool { return false }
	cp, err := g.validatePlayerAction(c)
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
		case p == nil:
			return ff
		case p.ID != cp.ID:
			return ff
		case g.PlayedCard == nil:
			return ff
		case ta == nil:
			return ff
		case g.PlayedCard.Kind == lampCard || g.PlayedCard.Kind == sLampCard:
			as := g.lampAreas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Kind == camelCard || g.PlayedCard.Kind == sCamelCard:
			as := g.camelAreas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Kind == swordCard:
			as := g.swordAreas(cp, ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Kind == carpetCard:
			as := g.carpetAreas()
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Kind == turbanCard && g.Stepped == 0:
			as := g.turban0Areas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Kind == turbanCard && g.Stepped == 1:
			as := g.turban1Areas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Kind == coinsCard:
			as := g.coinsAreas(ta)
			return func(a *Area) bool { return hasArea(as, a) }
		default:
			return ff
		}
	default:
		return ff
	}
}

func (g *Game) endOfTurnUpdateFor(p *Player) {
	if g.PlayedCard != nil {
		g.Jewels = *(g.PlayedCard)
	}

	for _, card := range p.Hand {
		card.FaceUp = true
	}
}

// IndexFor returns the index for the player and bool indicating whether player found.
func (g *Game) indexFor(p *Player) (int, bool) {
	if p == nil {
		return -1, false
	}

	for i, p2 := range g.Players {
		if p == p2 {
			return i, true
		}
	}
	return -1, false
}
