package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (cl *client) selectThiefHandler(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, cu, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	thiefArea, err := g.selectThief(c, cu)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	g, _, err = cl.putCachedGame(c, g, g.id(), g.Undo.Current)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	g.updateClickablesFor(cu, thiefArea)
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *Game) selectThief(c *gin.Context, cu *user.User) (*Area, error) {
	thiefArea, err := g.validateSelectThief(c, cu)
	if err != nil {
		return nil, err
	}

	g.thiefAreaID = thiefArea.areaID
	g.Phase = moveThiefPhase
	g.Undo.Update()

	g.appendEntry(message{
		"template": "select-thief",
		"area":     *thiefArea,
	})
	return thiefArea, nil
}

func (g *Game) validateSelectThief(c *gin.Context, cu *user.User) (*Area, error) {
	cp, err := g.validatePlayerAction(cu)
	if err != nil {
		return nil, err
	}

	a, err := g.getArea(c)
	switch {
	case err != nil:
		return nil, err
	case a == nil:
		return nil, fmt.Errorf("you must select an area: %w", sn.ErrValidation)
	case a.Thief != cp.ID:
		return nil, fmt.Errorf("you must select one of your thieves: %w", sn.ErrValidation)
	default:
		return a, nil
	}
}
