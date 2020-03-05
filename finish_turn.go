package got

import (
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

func finish(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		defer c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))

		g := gameFrom(c)
		switch g.Phase {
		case placeThieves:
			if err := g.placeThievesFinishTurn(c); err != nil {
				log.Errorf("g.placeThievesFinishTurn error: %v", err)
				return
			}

		case drawCard:
			if err := g.moveThiefFinishTurn(c); err != nil {
				log.Errorf("g.moveThiefFinishTurn error: %v", err)
				return
			}

		}
	}
}

func showPath(prefix string, sid string) string {
	return fmt.Sprintf("/%s/game/show/%s", prefix, sid)
}

func (g *Game) validateFinishTurn(c *gin.Context) (*stats.Stats, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")
	switch cp, s := g.CurrentPlayer(), stats.Fetched(c); {
	case s == nil:
		return nil, sn.NewVError("missing stats for player.")
	case !g.CUserIsCPlayerOrAdmin(c):
		return nil, sn.NewVError("Only the current player may finish a turn.")
	case !cp.PerformedAction:
		return nil, sn.NewVError("%s has yet to perform an action.", g.NameFor(cp))
	default:
		return s, nil
	}
}

// ps is an optional parameter.
// If no player is provided, assume current player.
func (g *Game) nextPlayer(ps ...game.Playerer) *Player {
	if nper := g.NextPlayerer(ps...); nper != nil {
		return nper.(*Player)
	}
	return nil
}

// ps is an optional parameter.
// If no player is provided, assume current player.
func (g *Game) previousPlayer(ps ...game.Playerer) *Player {
	if nper := g.PreviousPlayerer(ps...); nper != nil {
		return nper.(*Player)
	}
	return nil
}

func (g *Game) placeThievesNextPlayer(pers ...game.Playerer) (p *Player) {
	numThieves := 3
	if g.TwoThiefVariant {
		numThieves = 2
	}

	p = g.previousPlayer(pers...)

	if g.Round >= numThieves {
		p = nil
	} else if p.Equal(g.Players()[0]) {
		g.Round++
		p.beginningOfTurnReset()
	}
	return
}

func (g *Game) placeThievesFinishTurn(c *gin.Context) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")
	s, err := g.validatePlaceThievesFinishTurn(c)
	if err != nil {
		return err
	}

	oldCP := g.CurrentPlayer()
	np := g.placeThievesNextPlayer()
	if np == nil {
		g.SetCurrentPlayerers(g.Players()[0])
		g.CurrentPlayer().beginningOfTurnReset()
		g.startCardPlay(c)
	} else {
		g.SetCurrentPlayerers(np)
		np.beginningOfTurnReset()
	}

	newCP := g.CurrentPlayer()
	if newCP != nil && oldCP.ID() != newCP.ID() {
		g.SendTurnNotificationsTo(c, newCP)
	}

	s = s.GetUpdate(c, time.Time(g.UpdatedAt))
	ks := []*datastore.Key{s.Key}
	es := []interface{}{s}
	return g.saveWith(c, ks, es)
}

func (g *Game) validatePlaceThievesFinishTurn(c *gin.Context) (*stats.Stats, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")
	switch s, err := g.validateFinishTurn(c); {
	case err != nil:
		return nil, err
	case g.Phase != placeThieves:
		return nil, sn.NewVError("Expected %q phase but have %q phase.", placeThieves, g.Phase)
	default:
		return s, nil
	}
}

func (g *Game) moveThiefNextPlayer(pers ...game.Playerer) (np *Player) {
	cp := g.CurrentPlayer()
	g.endOfTurnUpdateFor(cp)
	ps := g.Players()
	np = g.nextPlayer(pers...)
	for !ps.allPassed() {
		if np.Passed {
			np = g.nextPlayer(np)
		} else {
			np.beginningOfTurnReset()
			return
		}
	}
	np = nil
	return
}

func (g *Game) moveThiefFinishTurn(c *gin.Context) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")
	s, err := g.validateMoveThiefFinishTurn(c)
	if err != nil {
		return err
	}

	oldCP := g.CurrentPlayer()
	np := g.moveThiefNextPlayer()

	// If no next player, end game
	if np == nil {
		g.finalClaim(c)
		ps := g.endGame(c)
		cs := contest.GenContests(c, ps)
		g.Status = game.Completed
		g.Phase = gameOver

		// Need to call SendTurnNotificationsTo before saving the new contests
		// SendEndGameNotifications relies on pulling the old contests from the db.
		// Saving the contests resulting in double counting.
		err = g.sendEndGameNotifications(c, ps, cs)
		if err != nil {
			// log but otherwise ignore send errors
			log.Warningf(err.Error())
		}

		s = s.GetUpdate(c, time.Time(g.UpdatedAt))
		l := len(cs)
		es := make([]interface{}, l)
		ks := make([]*datastore.Key, l)
		es[0] = s
		ks[0] = s.Key

		for i, c := range cs {
			es[i+1] = c
			ks[i+1] = c.Key
		}

		err = g.saveWith(c, ks, es)
		return err
	}

	// Otherwise, select next player and continue moving theives.
	g.SetCurrentPlayerers(np)
	if np.Equal(g.Players()[0]) {
		g.Turn++
	}
	g.Phase = playCard

	newCP := g.CurrentPlayer()
	if newCP != nil && oldCP.ID() != newCP.ID() {
		err = g.SendTurnNotificationsTo(c, newCP)
		if err != nil {
			// log but otherwise ignore send errors.
			log.Warningf(err.Error())
		}
	}
	s = s.GetUpdate(c, time.Time(g.UpdatedAt))
	ks := []*datastore.Key{s.Key}
	es := []interface{}{s}
	return g.saveWith(c, ks, es)
}

func (g *Game) validateMoveThiefFinishTurn(c *gin.Context) (*stats.Stats, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")
	switch s, err := g.validateFinishTurn(c); {
	case err != nil:
		return nil, err
	case g.Phase != drawCard:
		return nil, sn.NewVError(`Expected "Draw Card" phase but have %q phase.`, g.Phase)
	default:
		return s, nil
	}
}
