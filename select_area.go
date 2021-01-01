package got

import (
	"strconv"
	"strings"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (g *Game) selectArea(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	if err := g.validateSelectArea(c, cu); err != nil {
		return "got/flash_notice", game.None, err
	}

	switch cp := g.CurrentPlayer(); {
	case g.Admin == "admin-header":
		return "got/admin/header_dialog", game.Cache, nil
	case g.Admin == "admin-player-row-0":
		g.SelectedPlayerID = 0
		return "got/admin/player_dialog", game.Cache, nil
	case g.Admin == "admin-player-row-1":
		g.SelectedPlayerID = 1
		return "got/admin/player_dialog", game.Cache, nil
	case g.Admin == "admin-player-row-2":
		g.SelectedPlayerID = 2
		return "got/admin/player_dialog", game.Cache, nil
	case g.Admin == "admin-player-row-3":
		g.SelectedPlayerID = 3
		return "got/admin/player_dialog", game.Cache, nil
	case g.CanPlaceThief(cu, cp):
		template, err := g.placeThief(c, cu)
		return template, game.Cache, err
	case g.CanSelectCard(cu, cp):
		template, err := g.playCard(c, cu)
		return template, game.Cache, err
	case g.CanSelectThief(cu, cp):
		template, err := g.selectThief(c, cu)
		return template, game.Cache, err
	case g.CanMoveThief(cu, cp):
		template, err := g.moveThief(c, cu)
		return template, game.Cache, err
	default:
		return "got/flash_notice", game.None, sn.NewVError("Can't find action for selection.")
	}
}

func (g *Game) validateSelectArea(c *gin.Context, cu *user.User) (err error) {
	cp := g.CurrentPlayer()
	if !g.IsCurrentPlayer(cu) {
		err = sn.NewVError("Only the current player can perform an action.")
	} else if (cu != nil && !cu.Admin) && cp != nil && !g.CanPlaceThief(cu, cp) && !g.CanSelectCard(cu, cp) && !g.CanSelectThief(cu, cp) && !g.CanMoveThief(cu, cp) {
		err = sn.NewVError("You can't select an area right now.")
	}

	if err != nil {
		return
	}

	g.Admin = ""
	areaID := c.PostForm("area")
	switch splits := strings.Split(areaID, "-"); splits[0] {
	case "admin":
		g.Admin = areaID
	case "area":
		var row, col int
		if row, err = strconv.Atoi(splits[1]); err == nil {
			col, err = strconv.Atoi(splits[2])
		}

		switch {
		case err != nil:
		case row < rowA:
			err = sn.NewVError("Row too small")
		case row > rowG:
			err = sn.NewVError("Row too large")
		case g.NumPlayers == 2 && row > rowF:
			err = sn.NewVError("Row too large")
		case col < col1:
			err = sn.NewVError("Column too small")
		case col > col8:
			err = sn.NewVError("Column too large")
		default:
			g.SelectedAreaF = g.Grid[row][col]
		}
	case "card":
		if cardType := toCType(strings.TrimPrefix(areaID, "card-")); cardType == noType {
			err = sn.NewVError("Received invalid card type.")
		} else {
			for i, card := range cp.Hand {
				if card.Type == cardType {
					g.SelectedCardIndex = i
					return
				}
			}
			err = sn.NewVError("You don't have a %q card to play.", cardType)
		}
	default:
		err = sn.NewVError("Unable to determine selection.")
	}
	return
}
