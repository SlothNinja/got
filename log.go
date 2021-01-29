package got

import (
	"fmt"
	"html/template"
	"time"

	"github.com/SlothNinja/game"
)

// Entry stores information about a move in the game log.
type Entry struct {
	game.Entry
}

// GameLog stores entries of the game log.
type GameLog []Entryer

// Entryer specifies the interface for entries of the game log.
type Entryer interface {
	PhaseName() string
	Turn() int
	Round() int
	CreatedAt() time.Time
	HTML(g *Game) template.HTML
}

func (client *Client) newEntry() *Entry {
	e := new(Entry)
	e.PlayerID = game.NoPlayerID
	e.OtherPlayerID = game.NoPlayerID
	e.TurnF = client.Game.Turn
	e.PhaseF = client.Game.Phase
	e.SubPhaseF = client.Game.SubPhase
	e.RoundF = client.Game.Round
	e.CreatedAtF = time.Now()
	return e
}

func (client *Client) newEntryFor(p *Player) *Entry {
	e := client.newEntry()
	e.PlayerID = p.ID()
	return e
}

// PhaseName displays the turn and phase in an entry of the game log.
func (e *Entry) PhaseName() string {
	return fmt.Sprintf("Turn %d | Phase: %s", e.Turn(), phaseNames[e.Phase()])
}
