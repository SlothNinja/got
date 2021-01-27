package main

import (
	"sort"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

// player represents one of the players of the game.
type player struct {
	ID              int        `json:"id"`
	PerformedAction bool       `json:"performedAction"`
	Score           int        `json:"score"`
	Passed          bool       `json:"passed"`
	Colors          []sn.Color `json:"colors"`
	User            *sn.User   `json:"user"`
	Hand            Cards      `json:"hand"`
	DrawPile        Cards      `json:"drawPile"`
	DiscardPile     Cards      `json:"discardPile"`
	Stats           stats      `json:"stats"`
}

func (g *Game) pids() []int {
	pids := make([]int, len(g.players))
	for i := range g.players {
		pids[i] = g.players[i].ID
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
	sort.Slice(ps, func(i, j int) bool { return ps[i].compare(ps[j]) == sn.LessThan })
}

func reverse(ps []*player) {
	sort.Slice(ps, func(i, j int) bool { return false })
}

func (p *player) compare(p2 *player) sn.Comparison {
	byScore := p.compareByScore(p2)
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

func (p *player) compareByScore(p2 *player) sn.Comparison {
	switch {
	case p.Score < p2.Score:
		return sn.LessThan
	case p.Score > p2.Score:
		return sn.GreaterThan
	default:
		return sn.EqualTo
	}
}

func (p *player) compareByLamps(p2 *player) sn.Comparison {
	c0, c1 := p.Stats.Claimed[lampCard.String()], p2.Stats.Claimed[lampCard.String()]
	switch {
	case c0 < c1:
		return sn.LessThan
	case c0 > c1:
		return sn.GreaterThan
	default:
		return sn.EqualTo
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

func (p *player) compareByCamels(p2 *player) sn.Comparison {
	c0, c1 := p.Stats.Claimed[camelCard.String()], p2.Stats.Claimed[camelCard.String()]
	switch {
	case c0 < c1:
		return sn.LessThan
	case c0 > c1:
		return sn.GreaterThan
	default:
		return sn.EqualTo
	}
}

func (p *player) compareByCards(p2 *player) sn.Comparison {
	switch c0, c1 := len(p.Hand), len(p2.Hand); {
	case c0 < c1:
		return sn.LessThan
	case c0 > c1:
		return sn.GreaterThan
	}
	return sn.EqualTo
}

func (cl client) determinePlaces(c *gin.Context, g *Game) (sn.Places, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	// sort players by score
	sort.Slice(g.players, func(i, j int) bool { return g.players[i].compare(g.players[j]) != sn.LessThan })
	places := make(sn.Places, 0)
	rmap := make(sn.ResultsMap, 0)
	for i, p1 := range g.players {
		results := make(sn.Results, 0)
		tie := false
		for j, p2 := range g.players {
			r, err := cl.SN.GetRating(c, p2.User.Key, g.Type)
			if err != nil {
				log.Warningf(err.Error())
				return nil, err
			}
			result := &sn.Result{
				GameID: g.id(),
				Type:   g.Type,
				R:      r.R,
				RD:     r.RD,
			}
			switch c := p1.compare(p2); {
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
		} else if i == len(g.players)-1 {
			places = append(places, rmap)
		}
		p1.Stats.Finish = uint64(len(places))
	}
	return places, nil
}

func (g *Game) newPlayer(i int) *player {
	return &player{
		ID:          i + 1,
		Colors:      defaultColors()[:g.NumPlayers],
		Hand:        newStartHand(),
		DrawPile:    make(Cards, 0),
		DiscardPile: make(Cards, 0),
		User:        sn.ToUser(g.UserKeys[i], g.UserNames[i], g.UserEmailHashes[i]),
	}
}

func (p *player) beginningOfTurnReset() {
	p.PerformedAction = false
}

func (g *Game) endOfTurnUpdateFor(p *player) {
	if g.playedCard != nil {
		g.jewels = *(g.playedCard)
	}

	for _, card := range p.Hand {
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
