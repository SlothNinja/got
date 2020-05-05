package main

import (
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (g *Game) startSelectThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = selectThief
}

func (g *Game) selectThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validateSelectThief(c)
	if err != nil {
		return err
	}

	g.SelectedThiefAreaID = g.SelectedArea().areaID
	g.startMoveThief(c)
	return nil
}

func (g *Game) validateSelectThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	area, err := g.SelectedArea(), g.validatePlayerAction(c)
	switch {
	case err != nil:
		return err
	case area == nil:
		return fmt.Errorf("you must an area: %w", sn.ErrValidation)
	case area.Thief != g.CurrentPlayer().ID:
		return fmt.Errorf("you must select one of your thieves: %w", sn.ErrValidation)
	default:
		return nil
	}
}
