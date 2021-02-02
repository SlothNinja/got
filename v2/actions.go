package main

import (
	"errors"
	"fmt"

	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
)

func (cl *client) validatePlayerAction() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validateCPorAdmin()
	switch {
	case err != nil:
		return err
	case cl.cp.PerformedAction:
		return fmt.Errorf("current player already performed action: %w", sn.ErrValidation)
	default:
		return nil
	}
}

func (cl *client) validateCPorAdmin() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validateAdmin()
	if err == nil {
		cl.currentPlayer()
		return nil
	}

	return cl.validateCurrentPlayer()
}

func (cl *client) validateCurrentPlayer() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cp := cl.currentPlayer()
	switch {
	case cl.cu == nil:
		return user.ErrNotFound
	case cp == nil:
		return sn.ErrPlayerNotFound
	case cp.User.ID() != cl.cu.ID():
		return sn.ErrNotCurrentPlayer
	default:
		return nil
	}
}

func (cl *client) validateAdmin() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu := cl.CUser()
	switch {
	case cu == nil:
		return user.ErrNotFound
	case !cu.Admin:
		return errors.New("not admin")
	default:
		return nil
	}
}
