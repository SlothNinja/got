package main

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (cl client) pass(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	err = g.pass(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	ks, es := g.cache()
	_, err = cl.DS.Put(c, ks, es)
	if err != nil {
		sn.JErr(c, err)
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
	g.newEntryFor(cp.ID, message{"template": "pass"})
	return nil
}

func (g *Game) validatePass(c *gin.Context) (*player, error) {
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

func (cl client) passedFinishTurn(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	gcommitted, err := cl.getGCommited(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if gcommitted.Undo.Committed != g.Undo.Committed {
		sn.JErr(c, fmt.Errorf("invalid commit: %w", sn.ErrValidation))
		return
	}

	end, err := g.passedFinishTurn(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if end {
		cl.endGame(c, g)
		return
	}

	_, err = cl.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		g.Undo.Commit()
		_, err := tx.PutMulti(g.save())
		return err
	})
	if err != nil {
		sn.JErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"game": g})

}

func notPassed(p *player) bool { return !p.Passed }

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

func (g *Game) validatePassedFinishTurn(c *gin.Context) (*player, error) {
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
