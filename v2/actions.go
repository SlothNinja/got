package main

import (
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

func (g *Game) validatePlayerAction(c *gin.Context) (*Player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validateCPorAdmin(c)
	switch {
	case err != nil:
		return nil, err
	case cp.PerformedAction:
		return nil, fmt.Errorf("current player already performed action: %w", sn.ErrValidation)
	default:
		return cp, nil
	}
}

func (g *Game) validateCPorAdmin(c *gin.Context) (*Player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cu, err := g.validateAdmin(c)
	if err == nil {
		return g.currentPlayer(), nil
	}

	return g.validateCurrentPlayer(c, cu)
}

func (g *Game) validateCurrentPlayer(c *gin.Context, cu *user.User) (*Player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	log.Debugf("g.CPIDS: %#v", g.CPIDS)
	cp := g.currentPlayer()
	log.Debugf("cp.ID: %v", cp.ID)

	switch {
	case cu == nil:
		return nil, sn.ErrUserNotFound
	case cp == nil:
		return nil, sn.ErrPlayerNotFound
	case cp.User.ID() != cu.ID():
		log.Debugf("cp.User.ID(): %#v", cp.User.ID())
		log.Debugf("cu: %#v", cu.ID())
		return nil, sn.ErrNotCurrentPlayer
	default:
		return cp, nil
	}
}

func (g *Game) validateAdmin(c *gin.Context) (*user.User, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cu, err := user.FromSession(c)
	switch {
	case err != nil:
		return cu, err
	case cu == nil:
		return cu, sn.ErrUserNotFound
	case !cu.Admin:
		return cu, sn.ErrNotAdmin
	default:
		return cu, nil
	}
}
