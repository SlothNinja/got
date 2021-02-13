package main

import (
	"fmt"

	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
)

func (g *Game) validateFinishTurn(cu *user.User) (*player, error) {
	cp, err := g.validateCPorAdmin(cu)
	switch {
	case err != nil:
		return nil, err
	case !cp.PerformedAction:
		return nil, fmt.Errorf("%s has yet to perform an action: %w", cp.User.Name, sn.ErrValidation)
	default:
		return cp, nil
	}
}

type direction int

// ps is an optional parameter.
// If no player is provided, assume current player.
func (g *Game) nextPlayer(inc direction, p *player, tests ...func(*player) bool) *player {
	i, found := g.indexFor(p)
	if !found {
		return nil
	}

	for range g.players {
		i += int(inc)
		np := g.playerByIndex(i)
		if np.passed(tests...) {
			if i < 0 || i >= len(g.players) {
				g.Turn++
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
func (g *Game) playerByIndex(i int) *player {
	l := len(g.players)
	r := i % l
	if r < 0 {
		return g.players[l+r]
	}
	return g.players[r]
}

func (g *Game) lastPlayer() *player {
	l := len(g.players)
	if l == 0 {
		return nil
	}

	return g.players[l-1]
}

func (g *Game) firstPlayer() *player {
	if len(g.players) == 0 {
		return nil
	}

	return g.players[0]
}
