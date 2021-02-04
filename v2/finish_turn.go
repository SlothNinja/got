package main

import (
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
)

func (cl *client) validateFinishTurn() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	err := cl.validateCPorAdmin()
	switch {
	case err != nil:
		return err
	case !cl.cp.PerformedAction:
		return fmt.Errorf("%s has yet to perform an action: %w", cl.cp.User.Name, sn.ErrValidation)
	default:
		return nil
	}
}

type direction int

// ps is an optional parameter.
// If no player is provided, assume current player.
func (cl *client) nextPlayer(inc direction, p *player, tests ...func(*player) bool) *player {
	i, found := cl.g.indexFor(p)
	if !found {
		return nil
	}

	for range cl.g.players {
		i += int(inc)
		np := cl.playerByIndex(i)
		if np.passed(tests...) {
			if i < 0 || i >= len(cl.g.players) {
				cl.g.Turn++
			}
			return np
		}
	}
	return nil
}

func (p *player) passed(tests ...func(*player) bool) bool {
	for _, test := range tests {
		if !test(p) {
			return false
		}
	}
	return true
}

// implements ring buffer where index can be negative
func (cl *client) playerByIndex(i int) *player {
	if cl.g == nil {
		log.Warningf("cl.g is nil")
		return nil
	}

	l := len(cl.g.players)
	r := i % l
	if r < 0 {
		return cl.g.players[l+r]
	}
	return cl.g.players[r]
}

func (cl *client) lastPlayer() *player {
	if cl.g == nil {
		log.Warningf("cl.g is nil")
		return nil
	}

	l := len(cl.g.players)
	if l == 0 {
		return nil
	}

	return cl.g.players[l-1]
}

func (cl *client) firstPlayer() *player {
	if cl.g == nil {
		log.Warningf("cl.g is nil")
		return nil
	}

	if len(cl.g.players) == 0 {
		return nil
	}

	return cl.g.players[0]
}
