package main

import (
	"math/rand"

	"github.com/SlothNinja/game"
)

// Start begins a Guild of Thieves game.
func (cl *client) start() {
	cl.g.Status = game.Running
	cl.setupPhase()

	cl.g.Phase = placeThievesPhase
}

func (cl *client) addNewPlayers() {
	cl.g.players = make([]*player, cl.g.NumPlayers)
	for i := range cl.g.players {
		cl.g.players[i] = cl.newPlayer(i)
	}
}

func (cl *client) setupPhase() {
	cl.addNewPlayers()
	cl.randomTurnOrder()
	cl.createGrid()
	cp := cl.nextPlayer(backward, cl.g.players[0])
	cl.setCurrentPlayer(cp)

	cl.g.newEntry(message{
		"template": "start-game",
		"pids":     cl.pids(),
	})
	cl.g.Turn = 1
}

func (cl *client) randomTurnOrder() {
	rand.Shuffle(len(cl.g.players), func(i, j int) {
		cl.g.players[i], cl.g.players[j] = cl.g.players[j], cl.g.players[i]
	})
}
