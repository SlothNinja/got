package main

import (
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (cl client) placeThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := cl.getGame(c, 0)
	if err != nil {
		jerr(c, err)
		return
	}

	err = g.placeThief(c)
	if err != nil {
		jerr(c, err)
		return
	}

	ks, es := g.cache()
	_, err = cl.DS.Put(c, ks, es)
	if err != nil {
		jerr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *game) placeThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, a, err := g.validatePlaceThief(c)
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

func (g *game) validatePlaceThief(c *gin.Context) (*player, *Area, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validatePlayerAction(c)
	if err != nil {
		return nil, nil, err
	}

	a, err := g.getAreaFrom(c)
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

func (g *game) placeThievesNextPlayer(p *player) *player {
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

func (cl client) placeThievesFinishTurn(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := cl.getGame(c)
	if err != nil {
		jerr(c, err)
		return
	}

	gcommitted, err := cl.getGCommited(c)
	if err != nil {
		jerr(c, err)
		return
	}

	if gcommitted.Undo.Committed != g.Undo.Committed {
		jerr(c, fmt.Errorf("invalid commit: %w", sn.ErrValidation))
		return
	}

	cp, err := g.placeThievesFinishTurn(c)
	if err != nil {
		jerr(c, err)
		return
	}

	cp.Stats.Moves++
	cp.Stats.Think += time.Since(gcommitted.UpdatedAt)

	_, err = cl.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
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

func (g *game) placeThievesFinishTurn(c *gin.Context) (*player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validatePlaceThievesFinishTurn(c)
	if err != nil {
		return nil, err
	}

	np := g.placeThievesNextPlayer(cp)
	if np == nil {
		cp = g.firstPlayer()
		cp.beginningOfTurnReset()
		g.setCurrentPlayer(cp)
		g.Phase = playCardPhase
		return cp, nil
	}

	g.setCurrentPlayer(np)
	np.beginningOfTurnReset()
	if np != cp {
		g.SendTurnNotificationsTo(c, np)
	}

	return cp, nil
}

func (g *game) validatePlaceThievesFinishTurn(c *gin.Context) (*player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validateFinishTurn(c)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != placeThievesPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w",
			placeThievesPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, nil
	}
}
