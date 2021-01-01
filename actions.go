package got

import (
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
)

func (g *Game) validatePlayerAction(cu *user.User) error {
	if !g.IsCurrentPlayer(cu) {
		return sn.NewVError("Only the current player can perform an action.")
	}
	return nil
}

func (g *Game) validateAdminAction(cu *user.User) error {
	if cu == nil || !cu.Admin {
		return sn.NewVError("Only an admin can perform the selected action.")
	}
	return nil
}
