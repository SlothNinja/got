package main

import (
	"cloud.google.com/go/datastore"
)

type gcommitted struct{ *Game }

func (g *gcommitted) id() int64 {
	if g == nil || g.Key == nil {
		return 0
	}

	return g.Key.ID
}

// newGCommited creates a new Guild of Thieves game.
func newGCommited(id int64) *gcommitted { return &gcommitted{&Game{Key: newGCommittedKey(id)}} }

func newGCommittedKey(id int64) *datastore.Key {
	return datastore.IDKey(gCommitedKind, id, rootKey(id))
}

func (cl *client) setCurrentPlayer(p *player) {
	if cl.g == nil {
		cl.Log.Warningf("cl.g is nil")
		return
	}

	cl.g.CPIDS = nil
	if p == nil {
		return
	}
	cl.g.CPIDS = append(cl.g.CPIDS, p.ID)
	cl.currentPlayer()
}

func (cl *client) playerByID(id int) *player {
	if cl.g == nil {
		cl.Log.Warningf("cl.g was nil")
		return nil
	}

	for _, p := range cl.g.players {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}

func (cl *client) playerByUserID(id int64) *player {
	if cl.g == nil {
		cl.Log.Warningf("cl.g was nil")
		return nil
	}

	for i, uid := range cl.g.UserIDS {
		if uid == id {
			return cl.playerByID(i + 1)
		}
	}
	return nil
}

func (cl *client) playerByUserKey(key *datastore.Key) *player {
	if cl.g == nil {
		cl.Log.Warningf("cl.g was nil")
		return nil
	}

	for i, k := range cl.g.UserKeys {
		if k.Equal(key) {
			return cl.playerByID(i + 1)
		}
	}
	return nil
}

func (cl *client) playerByPID(pid int) *player {
	if cl.g == nil {
		cl.Log.Warningf("cl.g was nil")
		return nil
	}

	for _, p := range cl.g.players {
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
func (cl *client) currentPlayer() *player {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	if cl.cp != nil {
		return cl.cp
	}

	l := len(cl.g.CPIDS)
	if l != 1 {
		cl.cp = nil
		return cl.cp
	}

	pid := cl.g.CPIDS[0]
	for _, p := range cl.g.players {
		if p.ID == pid {
			cl.cp = p
			return cl.cp
		}
	}
	cl.cp = nil
	return cl.cp
}

func rootKey(id int64) *datastore.Key { return datastore.IDKey(rootKind, id, nil) }
