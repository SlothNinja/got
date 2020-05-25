package main

// State stores the game state.
type State struct {
	Players             []*Player `json:"players"`
	Grid                Grid      `json:"grid"`
	Jewels              Card      `json:"jewels"`
	Stepped             int       `json:"stepped"`
	PlayedCard          *Card     `json:"playedCard"`
	SelectedThiefAreaID areaID    `json:"selectedThiefAreaID"`
}
