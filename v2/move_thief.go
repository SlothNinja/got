package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (g *Game) startMoveThief(c *gin.Context) {
	g.Phase = moveThiefPhase
}

func (cl *client) moveThiefHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g, cu, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	err = g.moveThief(c, cu)
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

func (g *Game) moveThief(c *gin.Context, cu *user.User) error {
	cp, sa, ta, err := g.validateMoveThief(c, cu)
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
		g.updateClickablesFor(cu, sa)
	case g.playedCard.Kind == turbanCard && g.stepped == 1:
		g.stepped = 2
		g.claimItem(cp, ta)
		g.updateClickablesFor(cu, sa)
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
func (g *Game) validateMoveThief(c *gin.Context, cu *user.User) (*player, *Area, *Area, error) {
	sa, err := g.getArea(c)
	if err != nil {
		return nil, nil, nil, err
	}

	ta := g.selectedThiefArea()
	cp, err := g.validatePlayerAction(cu)
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

func (g *Game) bumpedTo(from, to *Area) *Area {
	switch {
	case from.Row > to.Row:
		return g.area(areaID{to.Row - 1, from.Column})
	case from.Row < to.Row:
		return g.area(areaID{to.Row + 1, from.Column})
	case from.Column > to.Column:
		return g.area(areaID{from.Row, to.Column - 1})
	case from.Column < to.Column:
		return g.area(areaID{from.Row, to.Column + 1})
	default:
		return nil
	}
}

func (cl *client) moveThiefFinishTurnHandler(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

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

	cp, np, err := g.moveThiefFinishTurn(cu)
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

	if cp != np {
		cl.sendTurnNotificationsTo(g, np)
	}
	c.JSON(http.StatusOK, gin.H{"game": g})

}

func (g *Game) moveThiefFinishTurn(cu *user.User) (*player, *player, error) {
	cp, err := g.validateMoveThiefFinishTurn(cu)
	if err != nil {
		return nil, nil, err
	}

	g.endOfTurnUpdate(cp)
	np := g.nextPlayer(forward, cp, notPassed)

	np.beginningOfTurnReset()
	g.setCurrentPlayer(np)
	g.Phase = playCardPhase

	return cp, np, nil
}

func (g *Game) validateMoveThiefFinishTurn(cu *user.User) (*player, error) {
	cp, err := g.validateFinishTurn(cu)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != moveThiefPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w", moveThiefPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, nil
	}
}
