package main

import (
	"errors"
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
)

func (g *Game) validatePlayerAction(cu *user.User) (*player, error) {
	cp, err := g.validateCPorAdmin(cu)
	switch {
	case err != nil:
		return nil, err
	case cp == nil:
		return nil, fmt.Errorf("not current player: %w", sn.ErrValidation)
	case cp.PerformedAction:
		return nil, fmt.Errorf("current player already performed action: %w", sn.ErrValidation)
	default:
		return cp, nil
	}
}

func (g *Game) validateCPorAdmin(cu *user.User) (*player, error) {
	cp, err := g.validateCurrentPlayer(cu)
	if err == nil {
		return cp, nil
	}

	err = validateAdmin(cu)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func (g *Game) validateCurrentPlayer(cu *user.User) (*player, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := g.currentPlayer()
	switch {
	case cu == nil:
		return nil, user.ErrNotFound
	case cp == nil:
		return nil, sn.ErrPlayerNotFound
	case g.uidFor(cp) != cu.ID():
		return nil, sn.ErrNotCurrentPlayer
	default:
		return cp, nil
	}
}

func validateAdmin(cu *user.User) error {
	switch {
	case cu == nil:
		return user.ErrNotFound
	case cu.Admin:
		return errors.New("not admin")
	default:
		return nil
	}
}
