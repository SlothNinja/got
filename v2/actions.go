package main

import (
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

func (g *game) validatePlayerAction(c *gin.Context) (*player, error) {
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

func (g *game) validateCPorAdmin(c *gin.Context) (*player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cu, err := g.validateAdmin(c)
	if err == nil {
		return g.currentPlayer(), nil
	}

	return g.validateCurrentPlayer(c, cu)
}

func (g *game) validateCurrentPlayer(c *gin.Context, cu *user.User) (*player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := g.currentPlayer()
	switch {
	case cu == nil:
		return nil, sn.ErrUserNotFound
	case cp == nil:
		return nil, sn.ErrPlayerNotFound
	case cp.User.ID() != cu.ID():
		return nil, sn.ErrNotCurrentPlayer
	default:
		return cp, nil
	}
}

func (g *game) validateAdmin(c *gin.Context) (*user.User, error) {
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
