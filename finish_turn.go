package got

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (client Client) finish(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		cu := user.CurrentFrom(c)

		s, err := client.Stats.ByUser(c, cu)
		if err != nil {
			log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
			return
		}

		var oldCP, newCP *Player
		g := gameFrom(c)
		switch g.Phase {
		case placeThieves:
			oldCP, newCP, err = g.placeThievesFinishTurn(c)
		case drawCard:
			oldCP, newCP, err = g.moveThiefFinishTurn(c)
		}

		if err != nil {
			log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
			return
		}

		var cs contest.Contests
		// newCP == nil => end game
		if newCP == nil {
			err = client.getCurrentRatings(c, g)
			if err != nil {
				log.Errorf(err.Error())
				restful.AddErrorf(c, err.Error())
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}

			g.finalClaim(c)
			ps := g.endGame(c)
			cs = contest.GenContests(c, ps)
			g.Status = game.Completed
			g.Phase = gameOver

			// Need to call SendEndGameNotifications before saving the new contests
			// SendEndGameNotifications relies on pulling the old contests from the db.
			// Saving the contests would result in double counting.
			err = client.sendEndGameNotifications(c, g, ps, cs)
			if err != nil {
				// log but otherwise ignore send errors
				log.Warningf(err.Error())
			}

		}

		if newCP != nil && oldCP.ID() != newCP.ID() {
			g.SendTurnNotificationsTo(c, newCP)
		}

		s = s.GetUpdate(c, g.UpdatedAt)
		ks, es := wrap(s, cs)
		err = client.saveWith(c, g, ks, es)
		if err != nil {
			log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
		}
		c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
	}
}

func (client Client) getCurrentRatings(c *gin.Context, g *Game) (err error) {
	for _, p := range g.Players() {
		p.Rating, err = client.Rating.GetRating(c, g.UserKeyFor(p), g.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func showPath(prefix string, sid string) string {
	return fmt.Sprintf("/%s/game/show/%s", prefix, sid)
}

func (g *Game) validateFinishTurn(c *gin.Context) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cp := g.CurrentPlayer()
	switch {
	case !g.CUserIsCPlayerOrAdmin(c):
		return sn.NewVError("Only the current player may finish a turn.")
	case !cp.PerformedAction:
		return sn.NewVError("%s has yet to perform an action.", g.NameFor(cp))
	default:
		return nil
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

func (g *Game) placeThievesFinishTurn(c *gin.Context) (*Player, *Player, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := g.validatePlaceThievesFinishTurn(c)
	if err != nil {
		return nil, nil, err
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
	return oldCP, newCP, nil
}

func (g *Game) validatePlaceThievesFinishTurn(c *gin.Context) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := g.validateFinishTurn(c)
	switch {
	case err != nil:
		return err
	case g.Phase != placeThieves:
		return sn.NewVError("Expected %q phase but have %q phase.", placeThieves, g.Phase)
	default:
		return nil
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

func (g *Game) moveThiefFinishTurn(c *gin.Context) (*Player, *Player, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := g.validateMoveThiefFinishTurn(c)
	if err != nil {
		return nil, nil, err
	}

	oldCP := g.CurrentPlayer()
	np := g.moveThiefNextPlayer()

	// If no next player, end game
	if np == nil {
		return oldCP, np, nil
	}

	// Otherwise, select next player and continue moving theives.
	g.SetCurrentPlayerers(np)
	if np.Equal(g.Players()[0]) {
		g.Turn++
	}
	g.Phase = playCard

	newCP := g.CurrentPlayer()
	return oldCP, newCP, nil
}

func (g *Game) validateMoveThiefFinishTurn(c *gin.Context) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := g.validateFinishTurn(c)
	switch {
	case err != nil:
		return err
	case g.Phase != drawCard:
		return sn.NewVError(`Expected "Draw Card" phase but have %q phase.`, g.Phase)
	default:
		return nil
	}
}
