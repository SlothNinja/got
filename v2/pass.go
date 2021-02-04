package main

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (cl *client) passHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	err := cl.getGame()
	if err != nil {
		cl.jerr(err)
		return
	}

	err = cl.pass()
	if err != nil {
		cl.jerr(err)
		return
	}

	ks, es := cl.g.cache()
	_, err = cl.DS.Put(c, ks, es)
	if err != nil {
		cl.jerr(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"game": cl.g})
}

func (cl *client) pass() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validatePass()
	if err != nil {
		return err
	}

	cl.cp.Passed = true
	cl.cp.PerformedAction = true
	cl.g.Phase = passedPhase

	cl.g.Undo.Update()
	cl.g.newEntryFor(cl.cp.ID, message{"template": "pass"})
	return nil
}

func (cl *client) validatePass() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validatePlayerAction()
	switch {
	case err != nil:
		return err
	case cl.g.Phase != playCardPhase:
		return fmt.Errorf("expected %q phase but have %q phase: %w", playCardPhase, cl.g.Phase, sn.ErrValidation)
	default:
		return nil
	}
}

func (cl *client) passedFinishTurnHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	err := cl.getGame()
	if err != nil {
		cl.jerr(err)
		return
	}

	err = cl.getGCommited()
	if err != nil {
		cl.jerr(err)
		return
	}

	if cl.gc.Undo.Committed != cl.g.Undo.Committed {
		cl.jerr(fmt.Errorf("invalid commit: %w", sn.ErrValidation))
		return
	}

	end, err := cl.passedFinishTurn()
	if err != nil {
		cl.jerr(err)
		return
	}

	if end {
		cl.endGame()
		return
	}

	_, err = cl.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		cl.g.Undo.Commit()
		_, err := tx.PutMulti(cl.g.save())
		return err
	})
	if err != nil {
		cl.jerr(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"game": cl.g})

}

func notPassed(p *player) bool { return !p.Passed }

func (cl *client) passedFinishTurn() (bool, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validatePassedFinishTurn()
	if err != nil {
		return false, err
	}

	cl.endOfTurnUpdate()
	np := cl.nextPlayer(forward, cl.cp, notPassed)

	// If no next player, end game
	if np == nil {
		return true, nil
	}

	// Otherwise, select next player and continue moving theives.
	np.beginningOfTurnReset()
	cl.setCurrentPlayer(np)
	cl.g.Phase = playCardPhase

	if np != cl.cp {
		err = cl.sendTurnNotificationsTo(np)
		if err != nil {
			// log but otherwise ignore send errors.
			log.Warningf(err.Error())
		}
	}
	return false, nil
}

func (cl *client) validatePassedFinishTurn() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validateFinishTurn()
	switch {
	case err != nil:
		return err
	case cl.g.Phase != passedPhase:
		return fmt.Errorf("expected %q phase but have %q phase: %w", passedPhase, cl.g.Phase, sn.ErrValidation)
	default:
		return nil
	}
}
