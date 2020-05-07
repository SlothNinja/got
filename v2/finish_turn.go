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

func (h *History) validateFinishTurn(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := h.CurrentPlayer()
	err := h.validateCPorAdmin(c)
	switch {
	case err != nil:
		return err
	case !cp.PerformedAction:
		return fmt.Errorf("%s has yet to perform an action: %w", cp.User.Name, sn.ErrValidation)
	default:
		return nil
	}
}

// ps is an optional parameter.
// If no player is provided, assume current player.
func (h *History) nextPlayer(inc int, ps ...*Player) *Player {
	var p *Player
	switch len(ps) {
	case 0:
		p = h.CurrentPlayer()
	case 1:
		p = ps[0]
	default:
		return nil
	}

	i, found := h.IndexFor(p)
	if !found {
		return nil
	}
	return h.playerByIndex(i + inc)
}

// // ps is an optional parameter.
// // If no player is provided, assume current player.
// func (g *Game) previousPlayer(ps ...*Player) *Player {
// 	var p *Player
// 	switch len(ps) {
// 	case 0:
// 		p = g.CurrentPlayer()
// 	case 1:
// 		p = ps[0]
// 	default:
// 		return nil
// 	}
//
// 	i, found := g.IndexFor(p)
// 	if !found {
// 		return nil
// 	}
// 	return g.playerByIndex(i - 1)
// }

// implements ring buffer where index can be negative
func (h *History) playerByIndex(i int) *Player {
	l := len(h.Players)
	r := i % l
	if r < 0 {
		return h.Players[l+r]
	}
	return h.Players[r]
}

func (h *History) placeThievesNextPlayer(ps ...*Player) *Player {
	numThieves := 3
	if h.TwoThiefVariant {
		numThieves = 2
	}

	p := h.nextPlayer(backward, ps...)
	switch {
	case h.Round >= numThieves:
		return nil
	case (p != nil) && (h.Players[0] != nil) && (p.ID == h.Players[0].ID):
		h.Round++
		p.beginningOfTurnReset()
		return p
	default:
		return p
	}
}

func (client Client) placeThievesFinishTurn(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	h, err := client.getHistory(c)
	if err != nil {
		jerr(c, err)
		return
	}

	g, err := client.getGame(c)
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

func (h *History) placeThievesFinishTurn(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := h.validatePlaceThievesFinishTurn(c)
	if err != nil {
		return err
	}

	oldCP := h.CurrentPlayer()
	np := h.placeThievesNextPlayer()
	if np == nil {
		h.setCurrentPlayer(h.Players[0])
		h.CurrentPlayer().beginningOfTurnReset()
		h.startCardPlay(c)
	} else {
		h.setCurrentPlayer(np)
		np.beginningOfTurnReset()
	}

	newCP := h.CurrentPlayer()
	if newCP != nil && oldCP != nil && oldCP.ID != newCP.ID {
		h.SendTurnNotificationsTo(c, newCP)
	}

	// zero flags
	h.SelectedPlayerID = 0
	h.BumpedPlayerID = 0
	h.SelectedAreaID = areaID{}
	h.SelectedCardIndex = 0
	h.Stepped = 0
	h.PlayedCard = nil
	h.JewelsPlayed = false
	h.SelectedThiefAreaID = areaID{}
	h.ClickAreas = nil
	h.Admin = ""

	return nil
}

func (h *History) validatePlaceThievesFinishTurn(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := h.validateFinishTurn(c)
	switch {
	case err != nil:
		return err
	case h.Phase != placeThieves:
		return fmt.Errorf("expected %q phase but have %q phase: %w",
			placeThieves, h.Phase, sn.ErrValidation)
	default:
		return nil
	}
}

func (h *History) moveThiefNextPlayer(ps ...*Player) *Player {
	cp := h.CurrentPlayer()
	h.endOfTurnUpdateFor(cp)
	np := h.nextPlayer(forward, ps...)
	for !allPassed(h.Players) {
		if np != nil && !np.Passed {
			np.beginningOfTurnReset()
			return np
		}
		np = h.nextPlayer(forward, np)
	}
	return nil
}

func (client Client) moveThiefFinishTurn(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	h, err := client.getHistory(c)
	if err != nil {
		jerr(c, err)
		return
	}

	g, err := client.getGame(c)
	if err != nil {
		jerr(c, err)
		return
	}

	if g.Undo.Committed != h.Undo.Committed {
		jerr(c, fmt.Errorf("invalid commit: %w", sn.ErrValidation))
		return
	}

	end, err := h.moveThiefFinishTurn(c)
	if err != nil {
		jerr(c, err)
		return
	}

	if end {
		h.finalClaim(c)
		ps, err := client.endGame(c, h)
		cs := sn.GenContests(c, ps)
		h.Status = sn.Completed
		h.Phase = gameOver

		// Need to call SendTurnNotificationsTo before saving the new contests
		// SendEndGameNotifications relies on pulling the old contests from the db.
		// Saving the contests resulting in double counting.
		err = client.sendEndGameNotifications(c, h, ps, cs)
		if err != nil {
			// log but otherwise ignore send errors
			log.Warningf(err.Error())
		}
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

func (h *History) moveThiefFinishTurn(c *gin.Context) (bool, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := h.validateMoveThiefFinishTurn(c)
	if err != nil {
		return false, err
	}

	oldCP := h.CurrentPlayer()
	np := h.moveThiefNextPlayer()

	// If no next player, end game
	if np == nil {
		return true, nil
	}

	// Otherwise, select next player and continue moving theives.
	h.setCurrentPlayer(np)
	if np != nil && h.Players[0] != nil && np.ID == h.Players[0].ID {
		h.Turn++
	}
	h.Phase = playCard

	newCP := h.CurrentPlayer()
	if newCP != nil && oldCP != nil && oldCP.ID != newCP.ID {
		err = h.SendTurnNotificationsTo(c, newCP)
		if err != nil {
			// log but otherwise ignore send errors.
			log.Warningf(err.Error())
		}
	}
	return false, nil
}

func (h *History) validateMoveThiefFinishTurn(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := h.validateFinishTurn(c)
	switch {
	case err != nil:
		return err
	case h.Phase != drawCard:
		return fmt.Errorf(`expected "Draw Card" phase but have %q phase: %w`, h.Phase, sn.ErrValidation)
	default:
		return nil
	}
}
