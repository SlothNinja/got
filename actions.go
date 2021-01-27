package got

import (
	"github.com/SlothNinja/sn"
)

func (client *Client) validatePlayerAction() error {
	if !client.Game.IsCurrentPlayer(client.CUser) {
		return sn.NewVError("Only the current player can perform an action.")
	}
	return nil
}

func (client *Client) validateAdminAction() error {
	if !client.CUser.IsAdmin() {
		return sn.NewVError("Only an admin can perform the selected action.")
	}
	return nil
}
