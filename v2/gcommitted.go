// Package got implements the card game, Guild of Thieves.
package main

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

// GCommited stores game state and header information.
type GCommited struct {
	Game
}

func (g *GCommited) ID() int64 {
	if g == nil || g.Key == nil {
		return 0
	}

	return g.Key.ID
}

// newGCommited creates a new Guild of Thieves game.
func newGCommited(id int64) *GCommited {
	g := new(GCommited)
	g.Key = newGCommittedKey(id)
	g.Type = sn.GOT
	return g
}

func newGCommittedKey(id int64) *datastore.Key {
	return datastore.IDKey(gCommitedKind, id, rootKey(id))
}

// func (g *GCommited) Load(ps []datastore.Property) error {
// 	err := datastore.LoadStruct(g, ps)
// 	if err != nil {
// 		return err
// 	}
//
// 	var s State
// 	err = json.Unmarshal([]byte(g.EncodedState), &s)
// 	if err != nil {
// 		return err
// 	}
// 	g.State = s
//
// 	var l Log
// 	err = json.Unmarshal([]byte(g.EncodedLog), &l)
// 	if err != nil {
// 		return err
// 	}
// 	g.Log = l
// 	return nil
// }
//
// func (g *GCommited) Save() ([]datastore.Property, error) {
//
// 	encodedState, err := json.Marshal(g.State)
// 	if err != nil {
// 		return nil, err
// 	}
// 	g.EncodedState = string(encodedState)
//
// 	encodedLog, err := json.Marshal(g.Log)
// 	if err != nil {
// 		return nil, err
// 	}
// 	g.EncodedLog = string(encodedLog)
//
// 	t := time.Now()
// 	if g.CreatedAt.IsZero() {
// 		g.CreatedAt = t
// 	}
//
// 	g.UpdatedAt = t
// 	return datastore.SaveStruct(g)
// }
//
// func (g *GCommited) LoadKey(k *datastore.Key) error {
// 	g.Key = k
// 	return nil
// }

func (g *Game) setCurrentPlayer(p *Player) {
	g.CPIDS = nil
	if p != nil {
		g.CPIDS = append(g.CPIDS, p.ID)
	}
}

// PlayerByID returns the player having the provided player id.
func (h *Game) PlayerByID(id int) *Player {
	if id <= 0 {
		return nil
	}

	for _, p := range h.Players {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}

// //func (g *History) PlayerBySID(sid string) (p *Player) {
// //	if per := g.Header.PlayerBySID(sid); per != nil {
// //		p = per.(*Player)
// //	}
// //	return
// //}

// PlayerByUserID returns the player having the user id.
func (g *Game) PlayerByUserID(id int64) *Player {
	if id <= 0 {
		return nil
	}

	for _, p := range g.Players {
		if p != nil && p.User != nil && p.User.ID() == id {
			return p
		}
	}
	return nil
}

//func (g *History) PlayerByIndex(index int) (player *Player) {
//	if p := g.PlayererByIndex(index); p != nil {
//		player = p.(*Player)
//	}
//	return
//}

func (g *Game) undoTurn(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	_, err := g.validateCPorAdmin(c)
	if err != nil {
		return err
	}

	return nil
}

// CurrentPlayer returns the player whose turn it is.
func (g *Game) currentPlayer() *Player {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	l := len(g.CPIDS)
	if l != 1 {
		return nil
	}

	pid := g.CPIDS[0]
	for _, p := range g.Players {
		if p.ID == pid {
			return p
		}
	}
	return nil
}

func rootKey(id int64) *datastore.Key {
	return datastore.IDKey(rootKind, id, nil)
}
