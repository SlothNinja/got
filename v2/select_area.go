package main

import (
	"fmt"

	"github.com/SlothNinja/sn"
)

func (cl *client) getArea() (*Area, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	obj := struct {
		ID areaID `json:"areaID"`
	}{}

	err := cl.ctx.ShouldBind(&obj)
	if err != nil {
		return nil, err
	}

	a := cl.area(obj.ID)
	if a == nil {
		return nil, fmt.Errorf("unable to find area: %w", sn.ErrValidation)
	}
	return a, nil
}

// func (g *History) selectArea(c *gin.Context) error {
// 	log.Debugf(msgEnter)
// 	defer log.Debugf(msgExit)
//
// 	err := g.validateSelectArea(c)
// 	if err != nil {
// 		return err
// 	}
//
// 	switch {
// 	case g.Admin == "admin-header":
// 		return nil
// 	case g.Admin == "admin-player-row-0":
// 		g.SelectedPlayerID = 0
// 		return nil
// 	case g.Admin == "admin-player-row-1":
// 		g.SelectedPlayerID = 1
// 		return nil
// 	case g.Admin == "admin-player-row-2":
// 		g.SelectedPlayerID = 2
// 		return nil
// 	case g.Admin == "admin-player-row-3":
// 		g.SelectedPlayerID = 3
// 		return nil
// 	case g.CanPlaceThief(c):
// 		return g.placeThief(c)
// 	case g.CanSelectCard(c):
// 		return g.playCard(c)
// 	case g.CanSelectThief(c):
// 		return g.selectThief(c)
// 	case g.CanMoveThief(c):
// 		return g.moveThief(c)
// 	default:
// 		return fmt.Errorf("can't find action for selection: %w", sn.ErrValidation)
// 	}
// }

// func (g *History) validateSelectArea(c *gin.Context) error {
//
// 	err := g.validateCPorAdmin(c)
// 	if err != nil {
// 		return err
// 	}
//
// 	if !g.CanPlaceThief(c) && !g.CanSelectCard(c) && !g.CanSelectThief(c) && !g.CanMoveThief(c) {
// 		return fmt.Errorf("you can't select an area right now: %w", sn.ErrValidation)
// 	}
//
// 	g.Admin = ""
// 	areaID := c.PostForm("area")
// 	splits := strings.Split(areaID, "-")
// 	switch splits[0] {
// 	case "admin":
// 		g.Admin = areaID
// 		return nil
// 	case "area":
// 		var row, col int
// 		if row, err = strconv.Atoi(splits[1]); err == nil {
// 			col, err = strconv.Atoi(splits[2])
// 		}
//
// 		switch {
// 		case err != nil:
// 			return err
// 		case row < rowA:
// 			return fmt.Errorf("row too small: %w", sn.ErrValidation)
// 		case row > rowG:
// 			return fmt.Errorf("row too large: %w", sn.ErrValidation)
// 		case g.NumPlayers == 2 && row > rowF:
// 			return fmt.Errorf("row too large: %w", sn.ErrValidation)
// 		case col < col1:
// 			return fmt.Errorf("column too small: %w", sn.ErrValidation)
// 		case col > col8:
// 			return fmt.Errorf("column too large: %w", sn.ErrValidation)
// 		default:
// 			g.SelectedAreaF = g.Grid[row][col]
// 			return nil
// 		}
// 	case "card":
// 		cardType := toCType(strings.TrimPrefix(areaID, "card-"))
// 		if cardType == noType {
// 			return fmt.Errorf("received invalid card type: %w", sn.ErrValidation)
// 		}
// 		cp := g.CurrentPlayer()
// 		for i, card := range cp.Hand {
// 			if card.Type == cardType {
// 				g.SelectedCardIndex = i
// 				return nil
// 			}
// 		}
// 		return fmt.Errorf("you don't have a %q card to play: %w", cardType, sn.ErrValidation)
// 	default:
// 		return fmt.Errorf("unable to determine selection: %w", sn.ErrValidation)
// 	}
// }
