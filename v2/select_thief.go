package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (h *History) startSelectThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	h.Phase = selectThief
}

func (client Client) selectThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	h, err := client.getHistory(c)
	if err != nil {
		jerr(c, err)
		return
	}

	err = h.selectThief(c)
	if err != nil {
		jerr(c, err)
		return
	}

	ks, es := h.cache()
	_, err = client.DS.Put(c, ks, es)
	if err != nil {
		jerr(c, err)
		return
	}

	h.updateClickablesFor(c, h.CurrentPlayer())
	c.JSON(http.StatusOK, gin.H{"game": h})
}

func (h *History) selectThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	a, err := h.validateSelectThief(c)
	if err != nil {
		return err
	}

	h.SelectedAreaID = a.areaID
	h.SelectedThiefAreaID = a.areaID
	h.startMoveThief(c)
	h.Undo.Update()
	return nil
}

func (h *History) validateSelectThief(c *gin.Context) (*Area, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := h.validatePlayerAction(c)
	if err != nil {
		return nil, err
	}

	a, err := h.getAreaFrom(c)
	switch {
	case err != nil:
		return nil, err
	case a == nil:
		return nil, fmt.Errorf("you must select an area: %w", sn.ErrValidation)
	case a.Thief != h.CurrentPlayer().ID:
		return nil, fmt.Errorf("you must select one of your thieves: %w", sn.ErrValidation)
	default:
		return a, nil
	}
}
