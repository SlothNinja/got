package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (cl *client) placeThiefHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g, cu, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	err = g.placeThief(c, cu)
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

func (g *Game) placeThief(c *gin.Context, cu *user.User) error {
	cp, a, err := g.validatePlaceThief(c, cu)
	if err != nil {
		return err
	}

	cp.PerformedAction = true
	cp.Score += a.Card.value()
	cp.Stats.Placed[g.Turn-1].inc(a.Card.Kind)
	a.Thief = cp.ID

	g.Undo.Update()

	g.newEntryFor(cp.ID, message{
		"template": "place-thief",
		"area":     *a,
	})
	return nil
}

func (g *Game) validatePlaceThief(c *gin.Context, cu *user.User) (*player, *Area, error) {
	cp, err := g.validatePlayerAction(cu)
	if err != nil {
		return nil, nil, err
	}

	a, err := g.getArea(c)
	if err != nil {
		return nil, nil, err
	}

	switch {
	case a == nil:
		return nil, nil, fmt.Errorf("you must select an area: %w", sn.ErrValidation)
	case a.Card == nil:
		return nil, nil, fmt.Errorf("you must select an area with a card: %w", sn.ErrValidation)
	case a.Thief != noPID:
		return nil, nil, fmt.Errorf("you must select an area without a thief: %w", sn.ErrValidation)
	default:
		return cp, a, nil
	}
}

func (g *Game) placeThievesNextPlayer(p *player) *player {
	numThieves := 3
	if g.TwoThiefVariant {
		numThieves = 2
	}

	p = g.nextPlayer(backward, p)
	if g.Turn > numThieves {
		return nil
	}
	return p
}

func (cl *client) placeThievesFinishTurnHandler(c *gin.Context) {
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

	cp, np, err := g.placeThievesFinishTurn(cu)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cp.Stats.Moves++
	cp.Stats.Think += time.Since(gc.UpdatedAt)

	err = cl.commit(c, g)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cl.sentRefreshMessages(c)

	if cp != np {
		cl.sendTurnNotificationsTo(g, np)
	}
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *Game) placeThievesFinishTurn(cu *user.User) (*player, *player, error) {
	cp, err := g.validatePlaceThievesFinishTurn(cu)
	if err != nil {
		return nil, nil, err
	}

	np := g.placeThievesNextPlayer(cp)
	if np == nil {
		cp = g.firstPlayer()
		cp.beginningOfTurnReset()
		g.setCurrentPlayer(cp)
		g.Phase = playCardPhase
		return cp, cp, nil
	}

	np.beginningOfTurnReset()
	if np != cp {
		g.setCurrentPlayer(np)
	}

	return cp, np, nil
}

func (g *Game) validatePlaceThievesFinishTurn(cu *user.User) (*player, error) {
	cp, err := g.validateFinishTurn(cu)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != placeThievesPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w", placeThievesPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, nil
	}
}
