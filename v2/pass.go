package main

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (client Client) pass(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := client.getGame(c)
	if err != nil {
		jerr(c, err)
		return
	}

	err = g.pass(c)
	if err != nil {
		jerr(c, err)
		return
	}

	ks, es := g.cache()
	_, err = client.DS.Put(c, ks, es)
	if err != nil {
		jerr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *Game) pass(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validatePass(c)
	if err != nil {
		return err
	}

	cp.Passed = true
	cp.PerformedAction = true
	g.Phase = passedPhase

	g.Undo.Update()
	g.newEntryFor(cp.ID, Message{"template": "pass"})
	return nil
}

func (g *Game) validatePass(c *gin.Context) (*Player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != playCardPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w",
			playCardPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, nil
	}
}

// func (g *Game) passedNextPlayer(p *Player) *Player {
// 	np := g.nextPlayer(forward, p)
// 	for !allPassed(g.Players) {
// 		if np != nil && !np.Passed {
// 			np.beginningOfTurnReset()
// 			return np
// 		}
// 		np = g.nextPlayer(forward, np)
// 	}
// 	return nil
// }

func (client Client) passedFinishTurn(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := client.getGame(c)
	if err != nil {
		jerr(c, err)
		return
	}

	gc, err := client.getGCommited(c)
	if err != nil {
		jerr(c, err)
		return
	}

	if gc.Undo.Committed != g.Undo.Committed {
		jerr(c, fmt.Errorf("invalid commit: %w", sn.ErrValidation))
		return
	}

	end, err := g.passedFinishTurn(c)
	if err != nil {
		jerr(c, err)
		return
	}

	if end {
		g.finalClaim(c)
		ps, err := client.endGame(c, g)
		cs := sn.GenContests(c, ps)
		g.Status = sn.Completed

		// Need to call SendTurnNotificationsTo before saving the new contests
		// SendEndGameNotifications relies on pulling the old contests from the db.
		// Saving the contests resulting in double counting.
		err = client.sendEndGameNotifications(c, g, ps, cs)
		if err != nil {
			// log but otherwise ignore send errors
			log.Warningf(err.Error())
		}
	}

	_, err = client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		g.Undo.Commit()
		_, err := tx.PutMulti(g.save())
		return err
	})
	if err != nil {
		jerr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"game": g})

}

func notPassed(p *Player) bool { return !p.Passed }

func (g *Game) passedFinishTurn(c *gin.Context) (bool, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validatePassedFinishTurn(c)
	if err != nil {
		return false, err
	}

	g.endOfTurnUpdateFor(cp)
	np := g.nextPlayer(forward, cp, notPassed)

	// If no next player, end game
	if np == nil {
		return true, nil
	}

	// Otherwise, select next player and continue moving theives.
	np.beginningOfTurnReset()
	g.setCurrentPlayer(np)
	g.Phase = playCardPhase

	if np != cp {
		err = g.SendTurnNotificationsTo(c, np)
		if err != nil {
			// log but otherwise ignore send errors.
			log.Warningf(err.Error())
		}
	}
	return false, nil
}

func (g *Game) validatePassedFinishTurn(c *gin.Context) (*Player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validateFinishTurn(c)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != passedPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w", passedPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, nil
	}
}
