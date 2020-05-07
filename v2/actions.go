package main

import (
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

func (h *History) validatePlayerAction(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := h.CurrentPlayer()
	err := h.validateCPorAdmin(c)

	switch {
	case err != nil:
		return err
	case cp.PerformedAction:
		return fmt.Errorf("current player already performed action: %w", sn.ErrValidation)
	default:
		return nil
	}
}

func (h *History) validateCPorAdmin(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cp := h.CurrentPlayer()
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
		log.Debugf("cp.User.ID: %v cu.ID: %v", cp.User.ID, cu.ID())
		return sn.ErrNotCPorAdmin
	default:
		return nil
	}
}

func (h *History) validateAdminAction(c *gin.Context) error {
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
