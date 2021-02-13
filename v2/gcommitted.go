package main

import (
	"cloud.google.com/go/datastore"
)

type gcommitted struct{ Game }

func (g *gcommitted) id() int64 {
	if g == nil || g.Key == nil {
		return 0
	}

	return g.Key.ID
}

// newGCommited creates a new Guild of Thieves game.
func newGCommited(id int64) *gcommitted { return &gcommitted{Game{Key: newGCommittedKey(id)}} }

func newGCommittedKey(id int64) *datastore.Key {
	return datastore.IDKey(gCommitedKind, id, rootKey(id))
}

func (g *Game) setCurrentPlayer(p *player) {
	g.CPIDS = nil
	if p == nil {
		return
	}
	g.CPIDS = append(g.CPIDS, p.ID)
}

func (g *Game) playerByID(id int) *player {
	for _, p := range g.players {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}

func (g *Game) playerByUserID(id int64) *player {
	for i, uid := range g.UserIDS {
		if uid == id {
			return g.playerByID(i + 1)
		}
	}
	return nil
}

func (g *Game) playerByUserKey(key *datastore.Key) *player {
	for i, k := range g.UserKeys {
		if k.Equal(key) {
			return g.playerByID(i + 1)
		}
	}
	return nil
}

func (g *Game) playerByPID(pid int) *player {
	for _, p := range g.players {
		if p != nil && p.ID == pid {
			return p
		}
	}
	return nil
}

// func (g *Game) undoTurn(c *gin.Context) error {
// 	log.Debugf(msgEnter)
// 	defer log.Debugf(msgExit)
//
// 	_, err := g.validateCPorAdmin(c)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }

// CurrentPlayer returns the player whose turn it is.
func (g *Game) currentPlayer() *player {
	l := len(g.CPIDS)
	if l != 1 {
		return nil
	}

	pid := g.CPIDS[0]
	return g.playerByPID(pid)
}

func rootKey(id int64) *datastore.Key { return datastore.IDKey(rootKind, id, nil) }
