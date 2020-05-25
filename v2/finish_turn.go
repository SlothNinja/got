package main

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

// func (client Client) finish(c *gin.Context) {
// 	log.Debugf(msgEnter)
// 	defer log.Debugf(msgExit)
//
// 	switch g.Phase {
// 	case placeThieves:
// 		client.placeThievesFinishTurn(c)
// 		return
// 	case drawCard:
// 		client.moveThiefFinishTurn(c)
// 		return
// 	}
//
// 	// // zero flags
// 	// g.SelectedPlayerID = 0
// 	// g.BumpedPlayerID = 0
// 	// g.SelectedAreaID = areaID{}
// 	// g.SelectedCardIndex = 0
// 	// g.Stepped = 0
// 	// g.PlayedCard = nil
// 	// g.JewelsPlayed = false
// 	// g.SelectedThiefAreaID = areaID{}
// 	// g.ClickAreas = nil
// 	// g.Admin = ""
//
// 	// if err != nil {
// 	// 	jerr(c, err)
// 	// 	return
// 	// }
//
// 	// err = client.saveWith(c, g, ks, es)
// 	// if err != nil {
// 	// 	jerr(c, err)
// 	// }
// 	// c.JSON(http.StatusOK, gin.H{"game": g})
// }

func (g *Game) validateFinishTurn(c *gin.Context) (*Player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validateCPorAdmin(c)
	switch {
	case err != nil:
		return nil, err
	case !cp.PerformedAction:
		return nil, fmt.Errorf("%s has yet to perform an action: %w", cp.User.Name, sn.ErrValidation)
	default:
		return cp, nil
	}
}

type direction int

// ps is an optional parameter.
// If no player is provided, assume current player.
func (g *Game) nextPlayer(inc direction, p *Player, tests ...func(*Player) bool) *Player {
	i, found := g.indexFor(p)
	if !found {
		return nil
	}

	for _ = range g.Players {
		i += int(inc)
		np := g.playerByIndex(i)
		if np.passed(tests...) {
			if i < 0 || i >= len(g.Players) {
				g.Turn++
			}
			return np
		}
	}
	return nil
	// if !found {
	// 	return nil
	// }

	// for j := 0; j < len(g.Players); j++ {
	// 	i += int(inc)
	// 	np := g.playerByIndex(i)
	// 	if passed(np, tests...) {
	// 		return np
	// 	}
	// }
	// return nil
}

func (p *Player) passed(tests ...func(*Player) bool) bool {
	for _, test := range tests {
		if !test(p) {
			return false
		}
	}
	return true
}

// implements ring buffer where index can be negative
func (g *Game) playerByIndex(i int) *Player {
	l := len(g.Players)
	r := i % l
	if r < 0 {
		return g.Players[l+r]
	}
	return g.Players[r]
}

func (g *Game) placeThievesNextPlayer(p *Player) *Player {
	numThieves := 3
	if g.TwoThiefVariant {
		numThieves = 2
	}

	p = g.nextPlayer(backward, p)
	if g.Turn > numThieves {
		return nil
	}
	return p
}

func (client Client) placeThievesFinishTurn(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	h, err := client.getGame(c)
	if err != nil {
		jerr(c, err)
		return
	}

	g, err := client.getGCommited(c)
	if err != nil {
		jerr(c, err)
		return
	}

	if g.Undo.Committed != h.Undo.Committed {
		jerr(c, fmt.Errorf("invalid commit: %w", sn.ErrValidation))
		return
	}

	err = h.placeThievesFinishTurn(c)
	if err != nil {
		jerr(c, err)
		return
	}

	_, err = client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		h.Undo.Commit()
		_, err := tx.PutMulti(h.save())
		return err
	})
	if err != nil {
		jerr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"game": h})
}

func (g *Game) lastPlayer() *Player {
	l := len(g.Players)
	if l == 0 {
		return nil
	}
	return g.Players[l-1]
}

func (g *Game) firstPlayer() *Player {
	if len(g.Players) == 0 {
		return nil
	}
	return g.Players[0]
}

func (g *Game) placeThievesFinishTurn(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validatePlaceThievesFinishTurn(c)
	if err != nil {
		return err
	}

	np := g.placeThievesNextPlayer(cp)
	if np == nil {
		cp = g.firstPlayer()
		cp.beginningOfTurnReset()
		g.setCurrentPlayer(cp)
		g.Phase = playCardPhase
		return nil
	}

	g.setCurrentPlayer(np)
	np.beginningOfTurnReset()
	if np != cp {
		g.SendTurnNotificationsTo(c, np)
	}

	return nil
}

func (g *Game) validatePlaceThievesFinishTurn(c *gin.Context) (*Player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validateFinishTurn(c)
	switch {
	case err != nil:
		return nil, err
	case g.Phase != placeThievesPhase:
		return nil, fmt.Errorf("expected %q phase but have %q phase: %w",
			placeThievesPhase, g.Phase, sn.ErrValidation)
	default:
		return cp, nil
	}
}
