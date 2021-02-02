package main

import (
	"sort"

	"github.com/SlothNinja/color"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/user"
)

// player represents one of the players of the game.
type player struct {
	ID              int           `json:"id"`
	PerformedAction bool          `json:"performedAction"`
	Score           int           `json:"score"`
	Passed          bool          `json:"passed"`
	Colors          []color.Color `json:"colors"`
	User            *user.User    `json:"user"`
	Hand            Cards         `json:"hand"`
	DrawPile        Cards         `json:"drawPile"`
	DiscardPile     Cards         `json:"discardPile"`
	Stats           stats         `json:"stats"`
}

func (cl *client) pids() []int {
	pids := make([]int, len(cl.g.players))
	for i := range cl.g.players {
		pids[i] = cl.g.players[i].ID
	}
	return pids
}

// Players is a slice of players of the game.
type Players []*player

func allPassed(ps []*player) bool {
	for _, p := range ps {
		if !p.Passed {
			return false
		}
	}
	return true
}

func byScore(ps []*player) {
	sort.Slice(ps, func(i, j int) bool { return ps[i].compare(ps[j]) == game.LessThan })
}

func reverse(ps []*player) {
	sort.Slice(ps, func(i, j int) bool { return false })
}

func (p *player) compare(p2 *player) game.Comparison {
	byScore := p.compareByScore(p2)
	if byScore != game.EqualTo {
		return byScore
	}

	byLamps := p.compareByLamps(p2)
	if byLamps != game.EqualTo {
		return byLamps
	}
	byCamels := p.compareByCamels(p2)
	if byCamels != game.EqualTo {
		return byCamels
	}
	return p.compareByCards(p2)
}

func (p *player) compareByScore(p2 *player) game.Comparison {
	switch {
	case p.Score < p2.Score:
		return game.LessThan
	case p.Score > p2.Score:
		return game.GreaterThan
	default:
		return game.EqualTo
	}
}

func (p *player) compareByLamps(p2 *player) game.Comparison {
	c0, c1 := p.Stats.Claimed[lampCard.String()], p2.Stats.Claimed[lampCard.String()]
	switch {
	case c0 < c1:
		return game.LessThan
	case c0 > c1:
		return game.GreaterThan
	default:
		return game.EqualTo
	}
}

func lampCount(cs []*Card) int {
	var count int
	for _, c := range cs {
		if c.Kind == lampCard || c.Kind == sLampCard {
			count++
		}
	}
	return count
}

func (p *player) compareByCamels(p2 *player) game.Comparison {
	c0, c1 := p.Stats.Claimed[camelCard.String()], p2.Stats.Claimed[camelCard.String()]
	switch {
	case c0 < c1:
		return game.LessThan
	case c0 > c1:
		return game.GreaterThan
	default:
		return game.EqualTo
	}
}

func (p *player) compareByCards(p2 *player) game.Comparison {
	switch c0, c1 := len(p.Hand), len(p2.Hand); {
	case c0 < c1:
		return game.LessThan
	case c0 > c1:
		return game.GreaterThan
	}
	return game.EqualTo
}

func (cl *client) determinePlaces() ([]contest.ResultsMap, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	// sort players by score
	sort.Slice(cl.g.players, func(i, j int) bool { return cl.g.players[i].compare(cl.g.players[j]) != game.LessThan })
	places := make([]contest.ResultsMap, 0)
	rmap := make(contest.ResultsMap, 0)
	for i, p1 := range cl.g.players {
		results := make([]*contest.Result, 0)
		tie := false
		for j, p2 := range cl.g.players {
			r, err := cl.Rating.Get(cl.ctx, p2.User.Key, cl.g.Type)
			if err != nil {
				log.Warningf(err.Error())
				return nil, err
			}
			result := &contest.Result{
				GameID: cl.g.id(),
				Type:   cl.g.Type,
				R:      r.R,
				RD:     r.RD,
			}
			switch c := p1.compare(p2); {
			case i == j:
			case c == game.GreaterThan:
				result.Outcome = 1
				results = append(results, result)
			case c == game.LessThan:
				result.Outcome = 0
				results = append(results, result)
			case c == game.EqualTo:
				result.Outcome = 0.5
				results = append(results, result)
				tie = true
			}
		}
		rmap[p1.User.Key] = results
		if !tie {
			places = append(places, rmap)
			rmap = make(contest.ResultsMap, 0)
		} else if i == len(cl.g.players)-1 {
			places = append(places, rmap)
		}
		p1.Stats.Finish = uint64(len(places))
	}
	return places, nil
}

func (cl *client) newPlayer(i int) *player {
	return &player{
		ID:          i + 1,
		Colors:      defaultColors()[:cl.g.NumPlayers],
		Hand:        newStartHand(),
		DrawPile:    make(Cards, 0),
		DiscardPile: make(Cards, 0),
	}
}

func (p *player) beginningOfTurnReset() {
	p.PerformedAction = false
}

func (cl *client) endOfTurnUpdate() {
	if cl.g.playedCard != nil {
		cl.g.jewels = *(cl.g.playedCard)
	}

	for _, card := range cl.cp.Hand {
		card.FaceUp = true
	}
}

// IndexFor returns the index for the player and bool indicating whether player found.
func (g *Game) indexFor(p *player) (int, bool) {
	if p == nil {
		return -1, false
	}

	for i, p2 := range g.players {
		if p == p2 {
			return i, true
		}
	}
	return -1, false
}

func (cl *client) emailFor(p *player) string {
	if cl.g == nil {
		cl.Log.Warningf("cl.g was nil")
		return ""
	}

	if p == nil {
		cl.Log.Warningf("p was nil")
		return ""
	}

	l, index := len(cl.g.UserEmails), p.ID-1
	if index >= 0 && index < l {
		return cl.g.UserEmails[index]
	}
	return ""
}

func (cl *client) nameFor(p *player) string {
	if cl.g == nil {
		cl.Log.Warningf("cl.g was nil")
		return ""
	}

	if p == nil {
		cl.Log.Warningf("p was nil")
		return ""
	}

	l, index := len(cl.g.UserNames), p.ID-1
	if index >= 0 && index < l {
		return cl.g.UserNames[index]
	}
	return ""
}
