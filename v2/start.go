package main

import (
	"math/rand"

	"github.com/SlothNinja/sn/v2"
)

// Start begins a Guild of Thieves game.
func (g *game) start() {
	g.Status = sn.Running
	g.setupPhase()

	g.Phase = placeThievesPhase
}

func (g *game) addNewPlayers() {
	g.players = make([]*player, g.NumPlayers)
	for i := range g.players {
		g.players[i] = g.newPlayer(i)
	}
}

func (g *game) setupPhase() {
	g.addNewPlayers()
	g.randomTurnOrder()
	g.createGrid()
	cp := g.nextPlayer(backward, g.players[0])
	g.setCurrentPlayer(cp)

	g.newEntry(message{
		"template": "start-game",
		"pids":     g.pids(),
	})
	g.Turn = 1
}

func (g *game) randomTurnOrder() {
	rand.Shuffle(len(g.players), func(i, j int) {
		g.players[i], g.players[j] = g.players[j], g.players[i]
	})
}
