package got

import (
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func (g *Game) validatePlayerAction(c *gin.Context) (err error) {
	if !g.CUserIsCPlayerOrAdmin(c) {
		err = sn.NewVError("Only the current player can perform an action.")
	}
	return
}

func (g *Game) validateAdminAction(c *gin.Context) (err error) {
	if !user.IsAdmin(c) {
		err = sn.NewVError("Only an admin can perform the selected action.")
	}
	return
}
