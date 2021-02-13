package main

import (
	"math/rand"

	"github.com/SlothNinja/game"
)

// Start begins a Guild of Thieves game.
func (g *Game) start() {
	g.Status = game.Running
	g.setupPhase()

	g.Phase = placeThievesPhase
}

func (g *Game) addNewPlayers() {
	g.players = make([]*player, g.NumPlayers)
	for i := range g.players {
		g.players[i] = g.newPlayer(i)
	}
}

func (g *Game) setupPhase() {
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

func (g *Game) randomTurnOrder() {
	rand.Shuffle(len(g.players), func(i, j int) {
		g.players[i], g.players[j] = g.players[j], g.players[i]
	})
}
