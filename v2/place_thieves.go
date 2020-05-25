package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (client Client) placeThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := client.getGame(c, 0)
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
	_, err = client.DS.Put(c, ks, es)
	if err != nil {
		jerr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *Game) placeThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, a, err := g.validatePlaceThief(c)
	if err != nil {
		return err
	}

	cp.PerformedAction = true
	cp.Score += a.Card.Value()
	a.Thief = cp.ID

	g.Undo.Update()

	g.newEntryFor(cp.ID, Message{
		"template": "place-thief",
		"area":     *a,
	})
	return nil
}

func (g *Game) validatePlaceThief(c *gin.Context) (*Player, *Area, error) {
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
