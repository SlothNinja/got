package got

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (client *Client) finish(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c
		g, cu := client.Game, client.CUser

		path := showPath(prefix, c.Param("hid"))
		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		s, err := client.User.StatsFor(c, cu)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		id, err := getID(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		err = client.getGameFor(id)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		var cs []*contest.Contest

		switch g.Phase {
		case placeThieves:
			err = g.placeThievesFinishTurn(c, cu)
		case drawCard:
			cs, err = client.moveThiefFinishTurn(c, g, cu)
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
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		err = client.saveWith(c, g, cu, s, cs)
		if err != nil {
			client.Log.Errorf(err.Error())
		}
		c.Redirect(http.StatusSeeOther, path)
	}
}

func showPath(prefix string, sid string) string {
	return fmt.Sprintf("/%s/game/show/%s", prefix, sid)
}

func (g *Game) validateFinishTurn(c *gin.Context, cu *user.User) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := g.CurrentPlayer()
	switch {
	case !g.IsCurrentPlayer(cu):
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

func (g *Game) placeThievesNextPlayer(cu *user.User, pers ...game.Playerer) (p *Player) {
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

func (g *Game) placeThievesFinishTurn(c *gin.Context, cu *user.User) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validatePlaceThievesFinishTurn(c, cu)
	if err != nil {
		return err
	}

	oldCP := g.CurrentPlayer()
	np := g.placeThievesNextPlayer(cu)
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

	return nil
}

func (g *Game) validatePlaceThievesFinishTurn(c *gin.Context, cu *user.User) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validateFinishTurn(c, cu)
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

func (client *Client) moveThiefFinishTurn(c *gin.Context, g *Game, cu *user.User) ([]*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := g.validateMoveThiefFinishTurn(c, cu)
	if err != nil {
		return nil, err
	}

	oldCP := g.CurrentPlayer()
	np := g.moveThiefNextPlayer()

	// If no next player, end game
	if np == nil {
		g.finalClaim(c)
		ps, err := client.endGame(c, g)
		cs := contest.GenContests(c, ps)
		g.Status = game.Completed
		g.Phase = gameOver

		// Need to call SendTurnNotificationsTo before saving the new contests
		// SendEndGameNotifications relies on pulling the old contests from the db.
		// Saving the contests resulting in double counting.
		err = client.sendEndGameNotifications(c, g, ps, cs)
		if err != nil {
			// log but otherwise ignore send errors
			client.Log.Warningf(err.Error())
		}

		return cs, nil
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
			client.Log.Warningf(err.Error())
		}
	}
	return nil, nil
}

func (g *Game) validateMoveThiefFinishTurn(c *gin.Context, cu *user.User) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validateFinishTurn(c, cu)
	switch {
	case err != nil:
		return err
	case g.Phase != drawCard:
		return sn.NewVError(`Expected "Draw Card" phase but have %q phase.`, g.Phase)
	default:
		return nil
	}
}
