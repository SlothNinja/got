package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (cl *client) passHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g, cu, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	err = g.pass(cu)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	g, _, err = cl.putCachedGame(c, g, g.id(), g.Undo.Current)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *Game) pass(cu *user.User) error {
	cp, err := g.validatePass(cu)
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

func (g *Game) validatePass(cu *user.User) (*player, error) {
	cp, err := g.validatePlayerAction(cu)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != playCardPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w", playCardPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, nil
	}
}

func (cl *client) passedFinishTurnHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g, cu, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	gc, err := cl.getGCommited(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if gc.Undo.Committed != g.Undo.Committed {
		sn.JErr(c, fmt.Errorf("invalid commit: %w", sn.ErrValidation))
		return
	}

	cp, np, err := g.passedFinishTurn(cu)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if np == nil {
		cl.endGame(c, g)
		return
	}

	err = cl.commit(c, g)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if cp != np {
		cl.sendTurnNotificationsTo(g, np)
	}
	c.JSON(http.StatusOK, gin.H{"game": g})

}

func notPassed(p *player) bool { return !p.Passed }

func (g *Game) passedFinishTurn(cu *user.User) (*player, *player, error) {
	cp, err := g.validatePassedFinishTurn(cu)
	if err != nil {
		return nil, nil, err
	}

	g.endOfTurnUpdate(cp)
	np := g.nextPlayer(forward, cp, notPassed)

	// If no next player, end game
	if np == nil {
		return cp, np, nil
	}

	// Otherwise, select next player and continue moving theives.
	np.beginningOfTurnReset()
	g.setCurrentPlayer(np)
	g.Phase = playCardPhase

	return cp, np, nil
}

func (g *Game) validatePassedFinishTurn(cu *user.User) (*player, error) {
	cp, err := g.validateFinishTurn(cu)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != passedPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w", passedPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, nil
	}
}
