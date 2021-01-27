package main

import (
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
)

func (g *Game) validatePlayerAction(cu *user.User) (*player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp, err := g.validateCPorAdmin(cu)
	switch {
	case err != nil:
		return nil, err
	case cp.PerformedAction:
		return nil, fmt.Errorf("current player already performed action: %w", sn.ErrValidation)
	default:
		return cp, nil
	}
}

func (g *Game) validateCPorAdmin(cu *user.User) (*player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validateAdmin(cu)
	if err == nil {
		return g.currentPlayer(), nil
	}

	return g.validateCurrentPlayer(cu)
}

func (g *Game) validateCurrentPlayer(cu *user.User) (*player, error) {
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

func (g *Game) validateAdmin(cu *user.User) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	if !cu.IsAdmin() {
		return sn.ErrNotAdmin
	}
	return nil
}
