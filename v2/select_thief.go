package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (cl client) selectThiefHandler(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cl.ctx = c

	err := cl.getGame()
	if err != nil {
		cl.jerr(err)
		return
	}

	thiefArea, err := cl.selectThief()
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

	cl.updateClickablesFor(cl.cp, thiefArea)
	c.JSON(http.StatusOK, gin.H{"game": cl.g})
}

func (cl *client) selectThief() (*Area, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	thiefArea, err := cl.validateSelectThief()
	if err != nil {
		return nil, err
	}

	cl.g.thiefAreaID = thiefArea.areaID
	cl.g.Phase = moveThiefPhase
	cl.g.Undo.Update()

	cl.g.appendEntry(message{
		"template": "select-thief",
		"area":     *thiefArea,
	})
	return thiefArea, nil
}

func (cl *client) validateSelectThief() (*Area, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validatePlayerAction()
	if err != nil {
		return nil, err
	}

	a, err := cl.getArea()
	switch {
	case err != nil:
		return nil, err
	case a == nil:
		return nil, fmt.Errorf("you must select an area: %w", sn.ErrValidation)
	case a.Thief != cl.cp.ID:
		return nil, fmt.Errorf("you must select one of your thieves: %w", sn.ErrValidation)
	default:
		return a, nil
	}
}
