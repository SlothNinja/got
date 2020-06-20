package main

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

type gcommitted struct{ game }

func (g *gcommitted) id() int64 {
	if g == nil || g.Key == nil {
		return 0
	}

	return g.Key.ID
}

// newGCommited creates a new Guild of Thieves game.
func newGCommited(id int64) *gcommitted { return &gcommitted{game{Key: newGCommittedKey(id)}} }

func newGCommittedKey(id int64) *datastore.Key { return datastore.IDKey(gCommitedKind, id, rootKey(id)) }

func (g *game) setCurrentPlayer(p *player) {
	g.CPIDS = nil
	if p != nil {
		g.CPIDS = append(g.CPIDS, p.ID)
	}
}

func (g *game) playerByID(id int) *player {
	if id <= 0 {
		return nil
	}

	for _, p := range g.players {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}

func (g *game) playerByUserID(id int64) *player {
	if id <= 0 {
		return nil
	}

	for _, p := range g.players {
		if p != nil && p.User != nil && p.User.ID() == id {
			return p
		}
	}
	return nil
}

func (g *game) playerByUserKey(key *datastore.Key) *player {
	if key == nil {
		return nil
	}

	for _, p := range g.players {
		if p != nil && p.User != nil && p.User.Key.Equal(key) {
			return p
		}
	}
	return nil
}

func (g *game) undoTurn(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	_, err := g.validateCPorAdmin(c)
	if err != nil {
		return err
	}

	return nil
}

// CurrentPlayer returns the player whose turn it is.
func (g *game) currentPlayer() *player {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	l := len(g.CPIDS)
	if l != 1 {
		return nil
	}

	pid := g.CPIDS[0]
	for _, p := range g.players {
		if p.ID == pid {
			return p
		}
	}
	return nil
}

func rootKey(id int64) *datastore.Key { return datastore.IDKey(rootKind, id, nil) }
