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

func (g *game) startMoveThief(c *gin.Context) {
	g.Phase = moveThiefPhase
}

func (cl client) moveThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	err = g.moveThief(c)
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

func (g *game) moveThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, sa, ta, err := g.validateMoveThief(c)
	if err != nil {
		return err
	}

	g.appendEntry(message{
		"template": "move-thief",
		"area":     *sa,
	})

	switch {
	case g.playedCard.Kind == swordCard:
		bumpedTo := g.bumpedTo(ta, sa)
		bumpedTo.Thief = sa.Thief
		g.appendEntry(message{
			"template": "bumped-thief",
			"area":     *bumpedTo,
		})
		bumpedPlayer := g.playerByID(bumpedTo.Thief)
		bumpedPlayer.Score += bumpedTo.Card.value() - sa.Card.value()
		g.claimItem(cp, ta)
		cp.PerformedAction = true
	case g.playedCard.Kind == turbanCard && g.stepped == 0:
		g.stepped = 1
		g.claimItem(cp, ta)
		g.thiefAreaID = sa.areaID
		g.updateClickablesFor(c, cp, sa)
	case g.playedCard.Kind == turbanCard && g.stepped == 1:
		g.stepped = 2
		g.claimItem(cp, ta)
		g.updateClickablesFor(c, cp, sa)
		cp.PerformedAction = true
	default:
		g.claimItem(cp, ta)
		cp.PerformedAction = true
	}
	sa.Thief = cp.ID
	cp.Score += sa.Card.value()

	g.Undo.Update()

	return nil
}

// return current player, selected area, thief area, and error
func (g *game) validateMoveThief(c *gin.Context) (*player, *Area, *Area, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	sa, err := g.getAreaFrom(c)
	if err != nil {
		return nil, nil, nil, err
	}

	ta := g.selectedThiefArea()
	cp, err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, nil, nil, err
	case sa == nil:
		return nil, nil, nil,
			fmt.Errorf("you must select a space to which to move your thief: %w", sn.ErrValidation)
	case ta == nil:
		return nil, nil, nil,
			fmt.Errorf("thief not selected: %w", sn.ErrValidation)
	case ta.Thief != cp.ID:
		return nil, nil, nil,
			fmt.Errorf("you must first select one of your thieves: %w", sn.ErrValidation)
	case (g.playedCard.Kind == lampCard || g.playedCard.Kind == sLampCard) && !hasArea(g.lampAreas(ta), sa):
		return nil, nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case (g.playedCard.Kind == camelCard || g.playedCard.Kind == sCamelCard) && !hasArea(g.camelAreas(ta), sa):
		return nil, nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.playedCard.Kind == coinsCard && !hasArea(g.coinsAreas(ta), sa):
		return nil, nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.playedCard.Kind == swordCard && !hasArea(g.swordAreas(cp, ta), sa):
		return nil, nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.playedCard.Kind == carpetCard && !g.isCarpetArea(sa):
		return nil, nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.playedCard.Kind == turbanCard && g.stepped == 0 && !hasArea(g.turban0Areas(ta), sa):
		return nil, nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.playedCard.Kind == turbanCard && g.stepped == 1 && !hasArea(g.turban1Areas(ta), sa):
		return nil, nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.playedCard.Kind == guardCard:
		return nil, nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	default:
		return cp, sa, ta, nil
	}
}

func (g *game) bumpedTo(from, to *Area) *Area {
	switch {
	case from.Row > to.Row:
		return g.getArea(areaID{to.Row - 1, from.Column})
	case from.Row < to.Row:
		return g.getArea(areaID{to.Row + 1, from.Column})
	case from.Column > to.Column:
		return g.getArea(areaID{from.Row, to.Column - 1})
	case from.Column < to.Column:
		return g.getArea(areaID{from.Row, to.Column + 1})
	default:
		return nil
	}
}

func (cl client) moveThiefFinishTurn(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	gcommited, err := cl.getGCommited(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if gcommited.Undo.Committed != g.Undo.Committed {
		sn.JErr(c, fmt.Errorf("invalid commit: %w", sn.ErrValidation))
		return
	}

	cp, send, err := g.moveThiefFinishTurn(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cp.Stats.Moves++
	cp.Stats.Think += time.Since(gcommited.UpdatedAt)

	_, err = cl.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		g.Undo.Commit()
		_, err := tx.PutMulti(g.save())
		return err
	})
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if send {
		err = g.SendTurnNotificationsTo(c, cp)
		if err != nil {
			// log but otherwise ignore send errors.
			log.Warningf(err.Error())
		}
	}
	c.JSON(http.StatusOK, gin.H{"game": g})

}

func (g *game) moveThiefFinishTurn(c *gin.Context) (*player, bool, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validateMoveThiefFinishTurn(c)
	if err != nil {
		return nil, false, err
	}

	g.endOfTurnUpdateFor(cp)
	np := g.nextPlayer(forward, cp, notPassed)

	np.beginningOfTurnReset()
	g.setCurrentPlayer(np)
	g.Phase = playCardPhase

	return np, np != cp, nil
}

func (g *game) validateMoveThiefFinishTurn(c *gin.Context) (*player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validateFinishTurn(c)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != moveThiefPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w",
			moveThiefPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, nil
	}
}
