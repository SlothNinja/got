package main

import (
	"html/template"
	"sort"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/schema"
	"github.com/SlothNinja/sn/v2"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

// Player represents one of the players of the game.
type Player struct {
	ID              int        `json:"id"`
	PerformedAction bool       `json:"performedAction"`
	Score           int        `json:"score"`
	Passed          bool       `json:"passed"`
	Colors          []sn.Color `json:"colors"`
	User            *User      `json:"user"`
	Log             GameLog    `json:"log"`
	Hand            Cards      `json:"hand"`
	DrawPile        Cards      `json:"drawPile"`
	DiscardPile     Cards      `json:"discardPile"`
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
func (cs Cards) CountFor(t cType) (faceUp, faceDown int) {
	for _, c := range cs {
		switch {
		case c.Type == t && c.FaceUp:
			faceUp++
		case c.Type == t && !c.FaceUp:
			faceDown++
		}
	}
	return
}

func lampCount(cs ...*Card) (count int) {
	for _, c := range cs {
		if c.Type == lamp || c.Type == sLamp {
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
		if c.Type == camel || c.Type == sCamel {
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

func newPlayer() *Player {
	p := &Player{
		Colors:      defaultColors(),
		Hand:        newStartHand(),
		DrawPile:    make(Cards, 0),
		DiscardPile: make(Cards, 0),
	}
	return p
}

func (g *Game) addNewPlayer(i int) {
	log.Debugf("g.Header: %#v", g.Header)
	log.Debugf("i: %d", i)
	p := newPlayer()
	g.Players = append(g.Players, p)
	p.ID = len(g.Players)
	p.User = toUser(g.UserKeys[i], g.UserNames[i], g.UserEmails[i])
}

// func createPlayer(g *Game, uid int64) *Player {
// 	p := newPlayer()
// 	p.ID = len(g.Players)
// 	return p
// }

func (p *Player) beginningOfTurnReset() {
	p.clearActions()
}

func (g *Game) beginningOfPhaseReset() {
	for _, p := range g.Players {
		p.clearActions()
		p.Passed = false
	}
}

func (p *Player) clearActions() {
	p.PerformedAction = false
	p.Log = make(GameLog, 0)
}

func (g *Game) updateClickablesFor(c *gin.Context, p *Player) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	canClick := g.CanClick(c, p)
	g.Grid.Each(func(a *Area) { a.Clickable = canClick(a) })
}

// CanClick a function specialized by current game context to test whether a player can click on
// a particular area in the grid.  The main benefit is the function provides a closure around area computions,
// essentially caching the results.
func (g *Game) CanClick(c *gin.Context, p *Player) func(*Area) bool {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	ff := func(a *Area) bool { return false }
	cp := g.CurrentPlayer()

	switch {
	case g == nil:
		return ff
	case cp == nil:
		return ff
	case g.validatePlayerAction(c) != nil:
		return ff
	case g.Phase == placeThieves:
		return func(a *Area) bool { return a.Thief == noPID }
	case g.Phase == selectThief:
		return func(a *Area) bool { return a.Thief == cp.ID }
	case g.Phase == moveThief:
		switch {
		case p == nil:
			return ff
		case p.ID != cp.ID:
			return ff
		case g.PlayedCard == nil:
			return ff
		case g.PlayedCard.Type == lamp || g.PlayedCard.Type == sLamp:
			as := g.lampAreas()
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Type == camel || g.PlayedCard.Type == sCamel:
			as := g.camelAreas()
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Type == sword:
			as := g.swordAreas()
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Type == carpet:
			as := g.camelAreas()
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Type == turban && g.Stepped == 0:
			as := g.turban0Areas()
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Type == turban && g.Stepped == 1:
			as := g.turban1Areas()
			return func(a *Area) bool { return hasArea(as, a) }
		case g.PlayedCard.Type == coins:
			as := g.coinsAreas()
			return func(a *Area) bool { return hasArea(as, a) }
		default:
			return ff
		}
	default:
		return ff
	}
}

// CanPlaceThief indicates whether a current player can place a thief.
func (g *Game) CanPlaceThief(c *gin.Context) bool {
	err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return false
	case g.Phase != placeThieves:
		return false
	default:
		return true
	}
}

// CanSelectCard indicates whether a current player can select a card to play.
func (g *Game) CanSelectCard(c *gin.Context) bool {
	err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return false
	case g.Phase != playCard:
		return false
	default:
		return true
	}
}

// CanSelectThief indicates whether current player can select a thief.
func (g *Game) CanSelectThief(c *gin.Context) bool {
	err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return false
	case g.Phase != selectThief:
		return false
	default:
		return true
	}
}

// CanMoveThief indicates whether current player can move a thief.
func (g *Game) CanMoveThief(c *gin.Context) bool {
	err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return false
	case g.Phase != moveThief:
		return false
	case g.SelectedThiefArea() == nil:
		return false
	default:
		return true
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

var playerValues = sslice{"Player.Passed", "Player.PerformedAction", "Player.Score"}

// func (g *Game) adminPlayer(c *gin.Context) (string, sn.ActionType, error) {
// 	log.Debugf(msgEnter)
// 	defer log.Debugf(msgExit)
//
// 	if err := g.adminUpdatePlayer(c, playerValues); err != nil {
// 		return "got/flash_notice", sn.None, err
// 	}
//
// 	return "", sn.Save, nil
// }

func (g *Game) adminUpdatePlayer(c *gin.Context, ss sslice) error {
	if err := g.validateAdminAction(c); err != nil {
		return err
	}

	p := g.selectedPlayer()
	values := make(map[string][]string)
	for _, key := range ss {
		if v := c.PostForm(key); v != "" {
			values[key] = []string{v}
		}
	}

	return schema.Decode(p, values)
}

func (g *Game) handMapFor(p *Player) (hm map[cType]int, count int) {
	hm = make(map[cType]int)
	for _, t := range g.cardTypes() {
		faceUp, faceDown := p.Hand.CountFor(t)
		if faceUp > 0 {
			hm[t] = faceUp
		}
		count += faceDown
	}
	return
}

// PlayCardDisplayFor outputs html for displaying a player's cards.
func (g *Game) PlayCardDisplayFor(p *Player) (s template.HTML) {
	cardTypes := 0
	hm, _ := g.handMapFor(p)
	for t, count := range hm {
		if count > 0 {
			cardTypes++
			pos := "push-right"
			if cardTypes%2 != 0 {
				s += restful.HTML("<div class='row' style='height:160px'>")
				pos = "pull-left"
			}

			name := t.IDString()
			s += restful.HTML("<div class=%q>", pos)
			s += restful.HTML("<div id='card-%s' data-tip=%q class='clickable card %s'></div>",
				name, t.toolTip(), name)
			s += restful.HTML("<div class='center'>%d</div></div>", count)

			if cardTypes%2 == 0 {
				s += restful.HTML("</div>")
			}
		}
	}
	if len(hm)%2 != 0 {
		s += restful.HTML("</div>")
	}
	return
}

// DisplayHandFor outputs html for displaying a player's hand.
func (g *Game) DisplayHandFor(c *gin.Context, p *Player) template.HTML {
	cu, err := user.FromSession(c)
	if err != nil {
		log.Warningf(err.Error())
		return ""
	}

	s := restful.HTML("<div id='player-hand-%d'>", p.ID)
	hm, faceDown := g.handMapFor(p)
	if cu.Admin || p.User.ID == cu.ID() || g.Phase == gameOver {
		for t, count := range hm {
			if count > 0 {
				name := t.IDString()
				s += restful.HTML("<div class='pull-left'>")
				s += restful.HTML("<div data-tip=%q class='card %s'></div>", t.toolTip(), name)
				s += restful.HTML("<div class='center'>%d</div></div>", count)
			}
		}
		if faceDown > 0 {
			s += restful.HTML("<div class='pull-left'>")
			s += restful.HTML("<div class='card card-back'></div>")
			s += restful.HTML("<div class='center'>%d</div></div>", faceDown)
		}
	} else {
		s += restful.HTML("<div class='pull-left'>")
		s += restful.HTML("<div class='card card-back'></div>")
		s += restful.HTML("<div class='center'>%d</div></div>", len(p.Hand))
	}
	s += restful.HTML("</div>")
	return s
}

// IndexFor returns the index for the player and bool indicating whether player found.
func (g *Game) IndexFor(p *Player) (int, bool) {
	if p == nil {
		return -1, false
	}

	for i, p2 := range g.Players {
		if p2 != nil && p.ID == p2.ID {
			return i, true
		}
	}
	return -1, false
}
