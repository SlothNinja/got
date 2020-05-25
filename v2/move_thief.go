package main

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (g *Game) startMoveThief(c *gin.Context) {
	g.Phase = moveThiefPhase
}

func (client Client) moveThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := client.getGame(c)
	if err != nil {
		jerr(c, err)
		return
	}

	err = g.moveThief(c)
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

func (g *Game) moveThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, sa, ta, err := g.validateMoveThief(c)
	if err != nil {
		return err
	}

	g.appendEntry(Message{
		"template": "move-thief",
		"area":     *sa,
	})

	switch {
	case g.PlayedCard.Kind == swordCard:
		bumpedTo := g.bumpedTo(ta, sa)
		bumpedTo.Thief = sa.Thief
		g.appendEntry(Message{
			"template": "bumped-thief",
			"area":     *bumpedTo,
		})
		bumpedPlayer := g.PlayerByID(bumpedTo.Thief)
		bumpedPlayer.Score += bumpedTo.Card.Value() - sa.Card.Value()
		g.claimItem(cp)
		cp.PerformedAction = true
	case g.PlayedCard.Kind == turbanCard && g.Stepped == 0:
		g.Stepped = 1
		g.claimItem(cp)
		g.SelectedThiefAreaID = sa.areaID
		g.updateClickablesFor(c, cp, sa)
	case g.PlayedCard.Kind == turbanCard && g.Stepped == 1:
		g.Stepped = 2
		g.claimItem(cp)
		g.updateClickablesFor(c, cp, sa)
		cp.PerformedAction = true
	default:
		g.claimItem(cp)
		cp.PerformedAction = true
	}
	sa.Thief = cp.ID
	cp.Score += sa.Card.Value()

	g.Undo.Update()

	return nil
}

// return current player, selected area, thief area, and error
func (g *Game) validateMoveThief(c *gin.Context) (*Player, *Area, *Area, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	sa, err := g.getAreaFrom(c)
	if err != nil {
		return nil, nil, nil, err
	}

	ta := g.SelectedThiefArea()
	cp, err := g.validatePlayerAction(c)
	switch {
	case err != nil:
		return nil, nil, nil, err
	case sa == nil:
		return nil, nil, nil, fmt.Errorf("you must select a space to which to move your thief: %w", sn.ErrValidation)
	case ta == nil:
		return nil, nil, nil, fmt.Errorf("thief not selected: %w", sn.ErrValidation)
	case ta.Thief != cp.ID:
		return nil, nil, nil, fmt.Errorf("you must first select one of your thieves: %w", sn.ErrValidation)
	case (g.PlayedCard.Kind == lampCard || g.PlayedCard.Kind == sLampCard) && !hasArea(g.lampAreas(ta), sa):
		return nil, nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case (g.PlayedCard.Kind == camelCard || g.PlayedCard.Kind == sCamelCard) && !hasArea(g.camelAreas(ta), sa):
		return nil, nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Kind == coinsCard && !hasArea(g.coinsAreas(ta), sa):
		return nil, nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Kind == swordCard && !hasArea(g.swordAreas(cp, ta), sa):
		return nil, nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Kind == carpetCard && !g.isCarpetArea(sa):
		return nil, nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Kind == turbanCard && g.Stepped == 0 && !hasArea(g.turban0Areas(ta), sa):
		return nil, nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Kind == turbanCard && g.Stepped == 1 && !hasArea(g.turban1Areas(ta), sa):
		return nil, nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	case g.PlayedCard.Kind == guardCard:
		return nil, nil, nil, fmt.Errorf("you can't move the selected thief to area %s: %w", sa.areaID, sn.ErrValidation)
	default:
		return cp, sa, ta, nil
	}
}

func (g *Game) bumpedTo(from, to *Area) *Area {
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

func (client Client) moveThiefFinishTurn(c *gin.Context) {
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

	cp, send, err := g.moveThiefFinishTurn(c)
	if err != nil {
		jerr(c, err)
		return
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

	if send {
		err = g.SendTurnNotificationsTo(c, cp)
		if err != nil {
			// log but otherwise ignore send errors.
			log.Warningf(err.Error())
		}
	}
	c.JSON(http.StatusOK, gin.H{"game": g})

}

func (g *Game) moveThiefFinishTurn(c *gin.Context) (*Player, bool, error) {
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

func (g *Game) validateMoveThiefFinishTurn(c *gin.Context) (*Player, error) {
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
