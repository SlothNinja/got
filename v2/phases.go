package main

import "encoding/json"

type Phase int

const (
	noPhase Phase = iota
	placeThievesPhase
	playCardPhase
	selectThiefPhase
	moveThiefPhase
	passedPhase
)

func (p Phase) String() string {
	return map[Phase]string{
		noPhase:           "None",
		placeThievesPhase: "Place Thieves",
		playCardPhase:     "Play Card",
		selectThiefPhase:  "Select Thief",
		moveThiefPhase:    "Move Thief",
		passedPhase:       "Passed",
	}[p]
}

func (p *Phase) fromString(s string) {
	*p = map[string]Phase{
		"None":          noPhase,
		"Place Thieves": placeThievesPhase,
		"Play Card":     playCardPhase,
		"Select Thief":  selectThiefPhase,
		"Move Thief":    moveThiefPhase,
		"Passed":        passedPhase,
	}[s]
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
	p.fromString(s)
	return nil
}
