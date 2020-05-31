package main

import "encoding/json"

type phase int

const (
	noPhase phase = iota
	placeThievesPhase
	playCardPhase
	selectThiefPhase
	moveThiefPhase
	passedPhase
)

func (p phase) String() string {
	return map[phase]string{
		noPhase:           "None",
		placeThievesPhase: "Place Thieves",
		playCardPhase:     "Play Card",
		selectThiefPhase:  "Select Thief",
		moveThiefPhase:    "Move Thief",
		passedPhase:       "Passed",
	}[p]
}

func (p *phase) fromString(s string) {
	*p = map[string]phase{
		"None":          noPhase,
		"Place Thieves": placeThievesPhase,
		"Play Card":     playCardPhase,
		"Select Thief":  selectThiefPhase,
		"Move Thief":    moveThiefPhase,
		"Passed":        passedPhase,
	}[s]
}

func (p phase) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *phase) UnmarshalJSON(bs []byte) error {
	var s string
	err := json.Unmarshal(bs, &s)
	if err != nil {
		return err
	}
	p.fromString(s)
	return nil
}
