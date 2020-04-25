package main

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

func (client Client) finish(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	var (
		ks  []*datastore.Key
		es  []interface{}
		err error
	)

	g := gameFrom(c)
	switch g.Phase {
	case placeThieves:
		ks, es, err = g.placeThievesFinishTurn(c)
	case drawCard:
		ks, es, err = client.moveThiefFinishTurn(c, g)
	}

	// zero flags
	g.SelectedPlayerID = 0
	g.BumpedPlayerID = 0
	g.SelectedAreaF = nil
	g.SelectedCardIndex = 0
	g.Stepped = 0
	g.PlayedCard = nil
	g.JewelsPlayed = false
	g.SelectedThiefAreaF = nil
	g.ClickAreas = nil
	g.Admin = ""

	if err != nil {
		log.Errorf(err.Error())
		c.Redirect(http.StatusSeeOther, showPath(c.Param("hid")))
		return
	}

	err = client.saveWith(c, g, ks, es)
	if err != nil {
		log.Errorf(err.Error())
	}
	c.Redirect(http.StatusSeeOther, showPath(c.Param("hid")))
}

func showPath(sid string) string {
	return fmt.Sprintf("/game/show/%s", sid)
}

func (g *Game) validateFinishTurn(c *gin.Context) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, s := g.CurrentPlayer(), user.FetchedStats(c)
	err := g.validateCPorAdmin(c)
	switch {
	case err != nil:
		return nil, err
	case s == nil:
		return nil, fmt.Errorf("missing stats for player: %w", sn.ErrValidation)
	case !cp.PerformedAction:
		return nil, fmt.Errorf("%s has yet to perform an action: %w", cp.User.Name, sn.ErrValidation)
	default:
		return s, nil
	}
}

// ps is an optional parameter.
// If no player is provided, assume current player.
func (g *Game) nextPlayer(inc int, ps ...*Player) *Player {
	var p *Player
	switch len(ps) {
	case 0:
		p = g.CurrentPlayer()
	case 1:
		p = ps[0]
	default:
		return nil
	}

	i, found := g.IndexFor(p)
	if !found {
		return nil
	}
	return g.playerByIndex(i + inc)
}

// // ps is an optional parameter.
// // If no player is provided, assume current player.
// func (g *Game) previousPlayer(ps ...*Player) *Player {
// 	var p *Player
// 	switch len(ps) {
// 	case 0:
// 		p = g.CurrentPlayer()
// 	case 1:
// 		p = ps[0]
// 	default:
// 		return nil
// 	}
//
// 	i, found := g.IndexFor(p)
// 	if !found {
// 		return nil
// 	}
// 	return g.playerByIndex(i - 1)
// }

// implements ring buffer where index can be negative
func (g *Game) playerByIndex(i int) *Player {
	l := len(g.Players)
	r := i % l
	if r < 0 {
		return g.Players[l+r]
	}
	return g.Players[r]
}

func (g *Game) placeThievesNextPlayer(ps ...*Player) *Player {
	numThieves := 3
	if g.TwoThiefVariant {
		numThieves = 2
	}

	p := g.nextPlayer(backward, ps...)
	switch {
	case g.Round >= numThieves:
		return nil
	case (p != nil) && (g.Players[0] != nil) && (p.ID == g.Players[0].ID):
		g.Round++
		p.beginningOfTurnReset()
		return p
	default:
		return p
	}
}

func (g *Game) placeThievesFinishTurn(c *gin.Context) ([]*datastore.Key, []interface{}, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)
	s, err := g.validatePlaceThievesFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	oldCP := g.CurrentPlayer()
	np := g.placeThievesNextPlayer()
	if np == nil {
		g.setCurrentPlayer(g.Players[0])
		g.CurrentPlayer().beginningOfTurnReset()
		g.startCardPlay(c)
	} else {
		g.setCurrentPlayer(np)
		np.beginningOfTurnReset()
	}

	newCP := g.CurrentPlayer()
	if newCP != nil && oldCP != nil && oldCP.ID != newCP.ID {
		g.SendTurnNotificationsTo(c, newCP)
	}

	s = s.GetUpdate(c, g.UpdatedAt)
	return []*datastore.Key{s.Key}, []interface{}{s}, nil
}

func (g *Game) validatePlaceThievesFinishTurn(c *gin.Context) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	s, err := g.validateFinishTurn(c)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != placeThieves:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w",
			placeThieves, g.Phase, sn.ErrValidation)
	default:
		return s, nil
	}
}

func (g *Game) moveThiefNextPlayer(ps ...*Player) *Player {
	cp := g.CurrentPlayer()
	g.endOfTurnUpdateFor(cp)
	np := g.nextPlayer(forward, ps...)
	for !allPassed(g.Players) {
		if np != nil && !np.Passed {
			np.beginningOfTurnReset()
			return np
		}
		np = g.nextPlayer(forward, np)
	}
	return nil
}

func (client Client) moveThiefFinishTurn(c *gin.Context, g *Game) ([]*datastore.Key, []interface{}, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)
	s, err := g.validateMoveThiefFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	oldCP := g.CurrentPlayer()
	np := g.moveThiefNextPlayer()

	// If no next player, end game
	if np == nil {
		g.finalClaim(c)
		ps, err := client.endGame(c, g)
		cs := sn.GenContests(c, ps)
		g.Status = sn.Completed
		g.Phase = gameOver

		// Need to call SendTurnNotificationsTo before saving the new contests
		// SendEndGameNotifications relies on pulling the old contests from the db.
		// Saving the contests resulting in double counting.
		err = client.sendEndGameNotifications(c, g, ps, cs)
		if err != nil {
			// log but otherwise ignore send errors
			log.Warningf(err.Error())
		}

		s = s.GetUpdate(c, g.UpdatedAt)
		ks, es := wrap(s, cs)
		return ks, es, nil
	}

	// Otherwise, select next player and continue moving theives.
	g.setCurrentPlayer(np)
	if np != nil && g.Players[0] != nil && np.ID == g.Players[0].ID {
		g.Turn++
	}
	g.Phase = playCard

	newCP := g.CurrentPlayer()
	if newCP != nil && oldCP != nil && oldCP.ID != newCP.ID {
		err = g.SendTurnNotificationsTo(c, newCP)
		if err != nil {
			// log but otherwise ignore send errors.
			log.Warningf(err.Error())
		}
	}
	s = s.GetUpdate(c, g.UpdatedAt)
	return []*datastore.Key{s.Key}, []interface{}{s}, nil
}

func (g *Game) validateMoveThiefFinishTurn(c *gin.Context) (*user.Stats, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	s, err := g.validateFinishTurn(c)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != drawCard:
		return nil, fmt.Errorf(`Expected "Draw Card" phase but have %q phase: %w`, g.Phase, sn.ErrValidation)
	default:
		return s, nil
	}
}
