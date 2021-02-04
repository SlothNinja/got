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

func (g *Game) startMoveThief(c *gin.Context) {
	g.Phase = moveThiefPhase
}

func (cl *client) moveThiefHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	err := cl.getGame()
	if err != nil {
		cl.jerr(err)
		return
	}

	err = cl.moveThief()
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

func (cl *client) moveThief() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	sa, ta, err := cl.validateMoveThief()
	if err != nil {
		return err
	}

	cl.g.appendEntry(message{
		"template": "move-thief",
		"area":     *sa,
	})

	switch {
	case cl.g.playedCard.Kind == swordCard:
		bumpedTo := cl.bumpedTo(ta, sa)
		bumpedTo.Thief = sa.Thief
		cl.g.appendEntry(message{
			"template": "bumped-thief",
			"area":     *bumpedTo,
		})
		bumpedPlayer := cl.playerByID(bumpedTo.Thief)
		bumpedPlayer.Score += bumpedTo.Card.value() - sa.Card.value()
		cl.g.claimItem(cl.cp, ta)
		cl.cp.PerformedAction = true
	case cl.g.playedCard.Kind == turbanCard && cl.g.stepped == 0:
		cl.g.stepped = 1
		cl.g.claimItem(cl.cp, ta)
		cl.g.thiefAreaID = sa.areaID
		cl.updateClickablesFor(cl.cp, sa)
	case cl.g.playedCard.Kind == turbanCard && cl.g.stepped == 1:
		cl.g.stepped = 2
		cl.g.claimItem(cl.cp, ta)
		cl.updateClickablesFor(cl.cp, sa)
		cl.cp.PerformedAction = true
	default:
		cl.g.claimItem(cl.cp, ta)
		cl.cp.PerformedAction = true
	}
	sa.Thief = cl.cp.ID
	cl.cp.Score += sa.Card.value()

	cl.g.Undo.Update()

	return nil
}

// return current player, selected area, thief area, and error
func (cl *client) validateMoveThief() (*Area, *Area, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	sa, err := cl.getArea()
	if err != nil {
		return nil, nil, err
	}

	ta := cl.selectedThiefArea()
	err = cl.validatePlayerAction()
	switch {
	case err != nil:
		return nil, nil, err
	case sa == nil:
		return nil, nil,
			fmt.Errorf("you must select a space to which to move your thief: %w", sn.ErrValidation)
	case ta == nil:
		return nil, nil,
			fmt.Errorf("thief not selected: %w", sn.ErrValidation)
	case ta.Thief != cl.cp.ID:
		return nil, nil,
			fmt.Errorf("you must first select one of your thieves: %w", sn.ErrValidation)
	case (cl.g.playedCard.Kind == lampCard || cl.g.playedCard.Kind == sLampCard) && !hasArea(cl.lampAreas(ta), sa):
		return nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case (cl.g.playedCard.Kind == camelCard || cl.g.playedCard.Kind == sCamelCard) && !hasArea(cl.camelAreas(ta), sa):
		return nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case cl.g.playedCard.Kind == coinsCard && !hasArea(cl.coinsAreas(ta), sa):
		return nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case cl.g.playedCard.Kind == swordCard && !hasArea(cl.swordAreas(cl.cp, ta), sa):
		return nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case cl.g.playedCard.Kind == carpetCard && !cl.isCarpetArea(sa):
		return nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case cl.g.playedCard.Kind == turbanCard && cl.g.stepped == 0 && !hasArea(cl.turban0Areas(ta), sa):
		return nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case cl.g.playedCard.Kind == turbanCard && cl.g.stepped == 1 && !hasArea(cl.turban1Areas(ta), sa):
		return nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case cl.g.playedCard.Kind == guardCard:
		return nil, nil,
			fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	default:
		return sa, ta, nil
	}
}

func (cl *client) bumpedTo(from, to *Area) *Area {
	switch {
	case from.Row > to.Row:
		return cl.area(areaID{to.Row - 1, from.Column})
	case from.Row < to.Row:
		return cl.area(areaID{to.Row + 1, from.Column})
	case from.Column > to.Column:
		return cl.area(areaID{from.Row, to.Column - 1})
	case from.Column < to.Column:
		return cl.area(areaID{from.Row, to.Column + 1})
	default:
		return nil
	}
}

func (cl client) moveThiefFinishTurnHandler(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

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

	np, send, err := cl.moveThiefFinishTurn()
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

	if send {
		err = cl.sendTurnNotificationsTo(np)
		if err != nil {
			// log but otherwise ignore send errors.
			cl.Log.Warningf(err.Error())
		}
	}
	c.JSON(http.StatusOK, gin.H{"game": cl.g})

}

func (cl *client) moveThiefFinishTurn() (*player, bool, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validateMoveThiefFinishTurn()
	if err != nil {
		return nil, false, err
	}

	cl.endOfTurnUpdate()
	np := cl.nextPlayer(forward, cl.cp, notPassed)

	np.beginningOfTurnReset()
	cl.setCurrentPlayer(np)
	cl.g.Phase = playCardPhase

	return np, np != cl.cp, nil
}

func (cl *client) validateMoveThiefFinishTurn() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validateFinishTurn()
	switch {
	case err != nil:
		return err
	case cl.g.Phase != moveThiefPhase:
		return fmt.Errorf("expected %q phase but have %q phase: %w", moveThiefPhase, cl.g.Phase, sn.ErrValidation)
	default:
		return nil
	}
}
