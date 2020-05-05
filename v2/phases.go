package main

import "encoding/json"

type Phase int

const (
	noPhase Phase = iota
	setup
	startGame
	placeThieves
	playCard
	selectThief
	moveThief
	claimItem
	drawCard
	finalClaim
	announceWinners
	gameOver
	endGame
	awaitPlayerInput
)

var phaseNames = map[Phase]string{
	noPhase:          "None",
	setup:            "Setup",
	startGame:        "Start Game",
	placeThieves:     "Place Thieves",
	playCard:         "Play Card",
	selectThief:      "Select Thief",
	moveThief:        "Move Thief",
	claimItem:        "Claim Magical Item",
	drawCard:         "Draw Card",
	finalClaim:       "Final Claim",
	announceWinners:  "Announce Winners",
	gameOver:         "Game Over",
	endGame:          "End Of Game",
	awaitPlayerInput: "Await Player Input",
}

var toPhase = map[string]Phase{
	"None":               noPhase,
	"Setup":              setup,
	"Start Game":         startGame,
	"Place Thieves":      placeThieves,
	"Play Card":          playCard,
	"Select Thief":       selectThief,
	"Move Thief":         moveThief,
	"Claim Magical Item": claimItem,
	"Draw Card":          drawCard,
	"Final Claim":        finalClaim,
	"Announce Winners":   announceWinners,
	"Game Over":          gameOver,
	"End Of Game":        endGame,
	"Await Player Input": awaitPlayerInput,
}

func (p Phase) String() string {
	return phaseNames[p]
}

func (p Phase) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Phase) UnmarshalJSON(bs []byte) error {
	var s string
	err := json.Unmarshal(bs, &s)
	if err != nil {
		return err
	}
	*p = toPhase[s]
	return nil
}

// // PhaseNames returns all phase names used by the game.
// func (g *Game) PhaseNames() sn.PhaseNameMap {
// 	return phaseNames
// }
//
// // PhaseName returns the name of the current phase of the game.
// func (g *Game) PhaseName() string {
// 	return phaseNames[g.Phase]
// }
