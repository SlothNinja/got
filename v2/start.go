package main

import (
	"math/rand"

	"github.com/SlothNinja/sn/v2"
)

// Start begins a Guild of Thieves game.
func (g *Game) start() {
	g.Status = sn.Running
	g.setupPhase()

	g.Phase = placeThievesPhase
}

func (g *Game) addNewPlayers() {
	g.Players = make([]*Player, g.NumPlayers)
	for i := range g.Players {
		g.Players[i] = g.newPlayer(i)
	}
}

func (g *Game) setupPhase() {
	g.addNewPlayers()
	g.randomTurnOrder()
	g.createGrid()
	cp := g.nextPlayer(backward, g.Players[0])
	g.setCurrentPlayer(cp)

	g.newEntry(Message{
		"template": "start-game",
		"pids":     g.pids(),
	})
	g.Turn = 1
}

func (g *Game) randomTurnOrder() {
	rand.Shuffle(len(g.Players), func(i, j int) {
		g.Players[i], g.Players[j] = g.Players[j], g.Players[i]
	})
}
