package main

import "encoding/json"

// state stores the game state.
type state struct {
	players     []*player
	grid        grid
	jewels      Card
	stepped     int
	playedCard  *Card
	thiefAreaID areaID
}

type jState struct {
	Players     []*player `json:"players"`
	Grid        grid      `json:"grid"`
	Jewels      Card      `json:"jewels"`
	Stepped     int       `json:"stepped"`
	PlayedCard  *Card     `json:"playedCard"`
	ThiefAreaID areaID    `json:"thiefAreaID"`
}

func (s state) MarshalJSON() ([]byte, error) {
	return json.Marshal(jState{
		Players:     s.players,
		Grid:        s.grid,
		Jewels:      s.jewels,
		Stepped:     s.stepped,
		PlayedCard:  s.playedCard,
		ThiefAreaID: s.thiefAreaID,
	})
}

func (s *state) UnmarshalJSON(bs []byte) error {
	var obj jState
	err := json.Unmarshal(bs, &obj)
	if err != nil {
		return err
	}
	s.players = obj.Players
	s.grid = obj.Grid
	s.jewels = obj.Jewels
	s.stepped = obj.Stepped
	s.playedCard = obj.PlayedCard
	s.thiefAreaID = obj.ThiefAreaID
	return nil
}
