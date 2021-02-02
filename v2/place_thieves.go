package main

import (
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (cl *client) placeThiefHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.getGame(0)
	if err != nil {
		cl.jerr(err)
		return
	}

	err = cl.placeThief()
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

func (cl *client) placeThief() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	a, err := cl.validatePlaceThief()
	if err != nil {
		return err
	}

	cl.cp.PerformedAction = true
	cl.cp.Score += a.Card.value()
	cl.cp.Stats.Placed[cl.g.Turn-1].inc(a.Card.Kind)
	a.Thief = cl.cp.ID

	cl.g.Undo.Update()

	cl.g.newEntryFor(cl.cp.ID, message{
		"template": "place-thief",
		"area":     *a,
	})
	return nil
}

func (cl *client) validatePlaceThief() (*Area, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validatePlayerAction()
	if err != nil {
		return nil, err
	}

	a, err := cl.getArea()
	if err != nil {
		return nil, err
	}

	switch {
	case a == nil:
		return nil, fmt.Errorf("you must select an area: %w", sn.ErrValidation)
	case a.Card == nil:
		return nil, fmt.Errorf("you must select an area with a card: %w", sn.ErrValidation)
	case a.Thief != noPID:
		return nil, fmt.Errorf("you must select an area without a thief: %w", sn.ErrValidation)
	default:
		return a, nil
	}
}

func (cl *client) placeThievesNextPlayer(p *player) *player {
	if cl.g == nil {
		cl.Log.Warningf("cl.g is nil")
		return nil
	}

	numThieves := 3
	if cl.g.TwoThiefVariant {
		numThieves = 2
	}

	p = cl.nextPlayer(backward, p)
	if cl.g.Turn > numThieves {
		return nil
	}
	return p
}

func (cl *client) placeThievesFinishTurnHandler(c *gin.Context) {
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

	err = cl.placeThievesFinishTurn()
	if err != nil {
		cl.jerr(err)
		return
	}

	cl.cp.Stats.Moves++
	cl.cp.Stats.Think += time.Since(cl.gc.UpdatedAt)

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

func (cl *client) placeThievesFinishTurn() error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := cl.validatePlaceThievesFinishTurn()
	if err != nil {
		return err
	}

	np := cl.placeThievesNextPlayer(cl.cp)
	if np == nil {
		cp := cl.firstPlayer()
		cp.beginningOfTurnReset()
		cl.setCurrentPlayer(cp)
		cl.g.Phase = playCardPhase
		return nil
	}

	np.beginningOfTurnReset()
	if np != cl.cp {
		cl.setCurrentPlayer(np)
		cl.sendTurnNotificationsTo(np)
	}

	return nil
}

func (cl *client) validatePlaceThievesFinishTurn() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validateFinishTurn()
	switch {
	case err != nil:
		return err
	case cl.g.Phase != placeThievesPhase:
		return fmt.Errorf("expected %q phase but have %q phase: %w", placeThievesPhase, cl.g.Phase, sn.ErrValidation)
	default:
		return nil
	}
}
