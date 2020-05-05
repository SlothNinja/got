package main

import (
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

func (g *Game) validatePlayerAction(c *gin.Context) error {
	cp := g.CurrentPlayer()
	err := g.validateCPorAdmin(c)

	switch {
	case err != nil:
		return err
	case cp == nil: // should never happen, if cp == nil => err != nil
		return sn.ErrPlayerNotFound
	case cp.PerformedAction:
		return fmt.Errorf("current player already performed action: %w", sn.ErrValidation)
	default:
		return nil
	}
}

func (g *Game) validateCPorAdmin(c *gin.Context) error {
	cp := g.CurrentPlayer()
	cu, err := user.FromSession(c)

	switch {
	case err != nil:
		return err
	case cu == nil:
		return sn.ErrUserNotFound
	case cu.Admin:
		return nil
	case cp == nil:
		return sn.ErrPlayerNotFound
	case cp.User.ID == 0:
		return sn.ErrNotCPorAdmin
	case cp.User.ID != cu.ID():
		log.Debugf("cp.User.ID: %d cu.ID: %d", cp.User.ID, cu.ID())
		return sn.ErrNotCPorAdmin
	default:
		return nil
	}
}

func (g *Game) validateAdminAction(c *gin.Context) error {
	cu, err := user.FromSession(c)
	switch {
	case err != nil:
		return err
	case !cu.Admin:
		return sn.ErrNotAdmin
	default:
		return nil
	}
}
