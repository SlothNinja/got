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
		path := showPath(prefix, c.Param("hid"))

		var err error
		client.CUser, err = client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		s, err := client.User.StatsFor(c, client.CUser)
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

		switch client.Game.Phase {
		case placeThieves:
			err = client.placeThievesFinishTurn()
		case drawCard:
			cs, err = client.moveThiefFinishTurn(c, client.Game, client.CUser)
		}

		// zero flags
		client.Game.SelectedPlayerID = 0
		client.Game.BumpedPlayerID = 0
		client.Game.SelectedAreaF = nil
		client.Game.SelectedCardIndex = 0
		client.Game.Stepped = 0
		client.Game.PlayedCard = nil
		client.Game.JewelsPlayed = false
		client.Game.SelectedThiefAreaF = nil
		client.Game.ClickAreas = nil
		client.Game.Admin = ""

		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		err = client.saveWith(c, client.Game, client.CUser, s, cs)
		if err != nil {
			client.Log.Errorf(err.Error())
		}
		c.Redirect(http.StatusSeeOther, path)
	}
}

func showPath(prefix string, sid string) string {
	return fmt.Sprintf("/%s/game/show/%s", prefix, sid)
}

func (client *Client) validateFinishTurn() error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := client.Game.CurrentPlayer()
	switch {
	case !client.Game.IsCurrentPlayer(client.CUser):
		return sn.NewVError("Only the current player may finish a turn.")
	case !cp.PerformedAction:
		return sn.NewVError("%s has yet to perform an action.", client.Game.NameFor(cp))
	default:
		return nil
	}
}

// ps is an optional parameter.
// If no player is provided, assume current player.
func (client *Client) nextPlayer(ps ...game.Playerer) *Player {
	nper := client.Game.NextPlayerer(ps...)
	if nper != nil {
		return nper.(*Player)
	}
	return nil
}

// ps is an optional parameter.
// If no player is provided, assume current player.
func (client *Client) previousPlayer(ps ...game.Playerer) *Player {
	nper := client.Game.PreviousPlayerer(ps...)
	if nper != nil {
		return nper.(*Player)
	}
	return nil
}

func (client *Client) placeThievesNextPlayer(pers ...game.Playerer) *Player {
	numThieves := 3
	if client.Game.TwoThiefVariant {
		numThieves = 2
	}

	p := client.previousPlayer(pers...)

	if client.Game.Round >= numThieves {
		return nil

	}

	if p.Equal(client.Game.Players()[0]) {
		client.Game.Round++
		p.beginningOfTurnReset()
	}
	return p
}

func (client *Client) placeThievesFinishTurn() error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.validatePlaceThievesFinishTurn()
	if err != nil {
		return err
	}

	oldCP := client.Game.CurrentPlayer()
	np := client.placeThievesNextPlayer()
	if np == nil {
		client.Game.SetCurrentPlayerers(client.Game.Players()[0])
		client.Game.CurrentPlayer().beginningOfTurnReset()
		client.startCardPlay()
	} else {
		client.Game.SetCurrentPlayerers(np)
		np.beginningOfTurnReset()
	}

	newCP := client.Game.CurrentPlayer()
	if newCP != nil && oldCP.ID() != newCP.ID() {
		client.Game.SendTurnNotificationsTo(client.Context, newCP)
	}

	return nil
}

func (client *Client) validatePlaceThievesFinishTurn() error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := client.validateFinishTurn()
	switch {
	case err != nil:
		return err
	case client.Game.Phase != placeThieves:
		return sn.NewVError("Expected %q phase but have %q phase.", placeThieves, client.Game.Phase)
	default:
		return nil
	}
}

func (client *Client) moveThiefNextPlayer(pers ...game.Playerer) *Player {
	cp := client.Game.CurrentPlayer()
	client.endOfTurnUpdateFor(cp)
	ps := client.Game.Players()
	np := client.nextPlayer(pers...)
	for !allPassed(ps) {
		if np.Passed {
			np = client.nextPlayer(np)
			continue
		}
		np.beginningOfTurnReset()
		return np
	}
	return nil
}

func (client *Client) moveThiefFinishTurn(c *gin.Context, g *Game, cu *user.User) ([]*contest.Contest, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.validateMoveThiefFinishTurn()
	if err != nil {
		return nil, err
	}

	oldCP := g.CurrentPlayer()
	np := client.moveThiefNextPlayer()

	// If no next player, end game
	if np == nil {
		client.finalClaim()
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

func (client *Client) validateMoveThiefFinishTurn() error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.validateFinishTurn()
	switch {
	case err != nil:
		return err
	case client.Game.Phase != drawCard:
		return sn.NewVError(`Expected "Draw Card" phase but have %q phase.`, client.Game.Phase)
	default:
		return nil
	}
}
