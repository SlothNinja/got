package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (cl client) selectThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := cl.getGame(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cp, thiefArea, err := g.selectThief(c)
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

	g.updateClickablesFor(c, cp, thiefArea)
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *game) selectThief(c *gin.Context) (*player, *Area, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, thiefArea, err := g.validateSelectThief(c)
	if err != nil {
		return nil, nil, err
	}

	g.thiefAreaID = thiefArea.areaID
	g.Phase = moveThiefPhase
	g.Undo.Update()

	g.appendEntry(message{
		"template": "select-thief",
		"area":     *thiefArea,
	})
	return cp, thiefArea, nil
}

func (g *game) validateSelectThief(c *gin.Context) (*player, *Area, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validatePlayerAction(c)
	if err != nil {
		return nil, nil, err
	}

	a, err := g.getAreaFrom(c)
	switch {
	case err != nil:
		return nil, nil, err
	case a == nil:
		return nil, nil, fmt.Errorf("you must select an area: %w", sn.ErrValidation)
	case a.Thief != cp.ID:
		return nil, nil, fmt.Errorf("you must select one of your thieves: %w", sn.ErrValidation)
	default:
		return cp, a, nil
	}
}
