package got

import (
	"strconv"
	"strings"

	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (client *Client) selectArea(c *gin.Context) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.validateSelectArea()
	if err != nil {
		client.flashError(err)
		return
	}

	cp := client.Game.CurrentPlayer()
	switch {
	// case g.Admin == "admin-header":
	// 	return "got/admin/header_dialog", game.Cache, nil
	// case g.Admin == "admin-player-row-0":
	// 	g.SelectedPlayerID = 0
	// 	return "got/admin/player_dialog", game.Cache, nil
	// case g.Admin == "admin-player-row-1":
	// 	g.SelectedPlayerID = 1
	// 	return "got/admin/player_dialog", game.Cache, nil
	// case g.Admin == "admin-player-row-2":
	// 	g.SelectedPlayerID = 2
	// 	return "got/admin/player_dialog", game.Cache, nil
	// case g.Admin == "admin-player-row-3":
	// 	g.SelectedPlayerID = 3
	// 	return "got/admin/player_dialog", game.Cache, nil
	case client.Game.CanPlaceThief(client.CUser, cp):
		client.placeThief()
	case client.Game.CanSelectCard(client.CUser, cp):
		client.playCard()
	case client.Game.CanSelectThief(client.CUser, cp):
		client.selectThief()
	case client.Game.CanMoveThief(client.CUser, cp):
		client.moveThief()
	default:
		client.flashError(sn.NewVError("can't find action for selection"))
	}
}

func (client *Client) validateSelectArea() error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	cp := client.Game.CurrentPlayer()
	if !client.Game.IsCurrentPlayer(client.CUser) {
		return sn.NewVError("only the current player can perform an action")
	}

	if !client.CUser.IsAdmin() && cp != nil && !client.Game.CanPlaceThief(client.CUser, cp) && !client.Game.CanSelectCard(client.CUser, cp) && !client.Game.CanSelectThief(client.CUser, cp) && !client.Game.CanMoveThief(client.CUser, cp) {
		return sn.NewVError("you can't select an area right now")
	}

	client.Game.Admin = ""
	areaID := client.Context.PostForm("area")
	switch splits := strings.Split(areaID, "-"); splits[0] {
	case "admin":
		client.Game.Admin = areaID
		return nil
	case "area":
		var row, col int
		row, err := strconv.Atoi(splits[1])
		if err == nil {
			col, err = strconv.Atoi(splits[2])
		}

		switch {
		case err != nil:
			return err
		case row < rowA:
			return sn.NewVError("Row too small")
		case row > rowG:
			return sn.NewVError("Row too large")
		case client.Game.NumPlayers == 2 && row > rowF:
			return sn.NewVError("Row too large")
		case col < col1:
			return sn.NewVError("Column too small")
		case col > col8:
			return sn.NewVError("Column too large")
		default:
			client.Game.SelectedAreaF = client.Game.Grid[row][col]
			return nil
		}
	case "card":
		cardType := toCType(strings.TrimPrefix(areaID, "card-"))
		if cardType == noType {
			return sn.NewVError("Received invalid card type.")
		}
		for i, card := range cp.Hand {
			if card.Type == cardType {
				client.Game.SelectedCardIndex = i
				return nil
			}
		}
		return sn.NewVError("You don't have a %q card to play.", cardType)
	default:
		return sn.NewVError("Unable to determine selection.")
	}
}
