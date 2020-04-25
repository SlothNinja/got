// Package got implements the card game, Guild of Thieves.
package main

import (
	"math/rand"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

// Game stores game state and header information.
type Game struct {
	*Header
	*State
}

// State stores the game state.
type State struct {
	Players            []*Player
	Log                GameLog
	Grid               grid
	Jewels             Card
	SelectedPlayerID   int
	BumpedPlayerID     int
	SelectedAreaF      *Area
	SelectedCardIndex  int
	Stepped            int
	PlayedCard         *Card
	JewelsPlayed       bool
	SelectedThiefAreaF *Area
	ClickAreas         areas
	Admin              string
}

// // GetPlayerers implements the GetPlayerers interfaces of the sn/games package.
// // Generally used to support common player manipulation functions of sn/games package.
// func (g *Game) GetPlayerers() game.Playerers {
// 	return g.Playerers
// }

// // Players returns a slice of player structs that store various information about each player.
// func (g *Game) Players() (ps Players) {
// 	pers := g.GetPlayerers()
// 	length := len(pers)
// 	if length > 0 {
// 		ps = make(Players, length)
// 		for i, p := range pers {
// 			ps[i] = p.(*Player)
// 		}
// 	}
// 	return
// }
//
// func (g *Game) setPlayers(ps Players) {
// 	length := len(ps)
// 	if length > 0 {
// 		pers := make(game.Playerers, length)
// 		for i, p := range ps {
// 			pers[i] = p
// 		}
// 		g.Playerers = pers
// 	}
// }

// Games is a slice of Guild of Thieves games.
// type Games []*Game

// Start begins a Guild of Thieves game.
func (g *Game) Start(c *gin.Context) error {
	g.Status = sn.Running
	return g.setupPhase(c)
}

func (g *Game) addNewPlayers() {
	for i := range g.UserKeys {
		g.addNewPlayer(i)
	}
}

func (g *Game) setupPhase(c *gin.Context) error {
	g.Turn = 0
	g.Phase = setup
	g.addNewPlayers()
	g.randomTurnOrder()
	g.createGrid()
	// for _, p := range g.Players {
	// 	g.newSetupEntryFor(p)
	// }
	cp := g.nextPlayer(backward, g.Players[0])
	g.setCurrentPlayer(cp)
	g.beginningOfPhaseReset()
	return g.start(c)
}

// type setupEntry struct {
// 	*Entry
// }
//
// func (g *Game) newSetupEntryFor(p *Player) (e *setupEntry) {
// 	e = new(setupEntry)
// 	e.Entry = g.newEntryFor(p)
// 	p.Log = append(p.Log, e)
// 	g.Log = append(g.Log, e)
// 	return
// }
//
// func (e *setupEntry) HTML(g *Game) template.HTML {
// 	return restful.HTML("%s received 2 lamps and 1 camel.", g.NameByPID(e.PlayerID))
// }

func (g *Game) start(c *gin.Context) error {
	g.Phase = startGame
	// g.newStartEntry()
	return g.placeThieves(c)
}

// type startEntry struct {
// 	*Entry
// }
//
// func (g *Game) newStartEntry() *startEntry {
// 	e := new(startEntry)
// 	e.Entry = g.newEntry()
// 	g.Log = append(g.Log, e)
// 	return e
// }
//
// func (e *startEntry) HTML(g *Game) template.HTML {
// 	names := make([]string, g.NumPlayers)
// 	for i, p := range g.Players() {
// 		names[i] = g.NameFor(p)
// 	}
// 	return restful.HTML("Good luck %s.  Have fun.", restful.ToSentence(names))
// }

func (g *Game) setCurrentPlayer(p *Player) {
	g.CPUserIndices = nil
	if p != nil {
		g.CPUserIndices = append(g.CPUserIndices, p.ID)
	}
}

// PlayerByID returns the player having the provided player id.
func (g *Game) PlayerByID(id int) *Player {
	if id <= 0 {
		return nil
	}

	for _, p := range g.Players {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}

// //func (g *Game) PlayerBySID(sid string) (p *Player) {
// //	if per := g.Header.PlayerBySID(sid); per != nil {
// //		p = per.(*Player)
// //	}
// //	return
// //}

// PlayerByUserID returns the player having the user id.
func (g *Game) PlayerByUserID(id int64) *Player {
	if id <= 0 {
		return nil
	}

	for _, p := range g.Players {
		if p != nil && p.User.ID() == id {
			return p
		}
	}
	return nil
}

//func (g *Game) PlayerByIndex(index int) (player *Player) {
//	if p := g.PlayererByIndex(index); p != nil {
//		player = p.(*Player)
//	}
//	return
//}

func (g *Game) undoTurn(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validateCPorAdmin(c)
	if err != nil {
		return err
	}

	return nil
}

// CurrentPlayer returns the player whose turn it is.
func (g *Game) CurrentPlayer() *Player {
	l := len(g.CPUserIndices)
	if l != 1 {
		return nil
	}
	i := g.CPUserIndices[0]
	l = len(g.Players)
	if i >= l {
		return nil
	}
	return g.Players[i]
}

// Convenience method for conditionally logging Debug information
// based on package global const debug
//const debug = true
//
//func (g *Game) debugf(format string, args ...interface{}) {
//	if debug {
//		g.Debugf(format, args...)
//	}
//}

type sslice []string

func (ss sslice) include(s string) bool {
	for _, str := range ss {
		if str == s {
			return true
		}
	}
	return false
}

var headerValues = sslice{
	"Header.Title",
	"Header.Turn",
	"Header.Phase",
	"Header.Round",
	"Header.Password",
	"Header.CPUserIndices",
	"Header.WinnerIDS",
	"Header.Status",
}

// func (g *Game) adminHeader(c *gin.Context) (string, game.ActionType, error) {
// 	log.Debugf(msgEnter)
// 	defer log.Debugf(msgExit)
//
// 	if err := g.adminUpdateHeader(c, headerValues); err != nil {
// 		return "got/flash_notice", game.None, err
// 	}
//
// 	return "", game.Save, nil
// }
//
// func (g *Game) adminUpdateHeader(c *gin.Context, ss sslice) error {
// 	if err := g.validateAdminAction(c); err != nil {
// 		return err
// 	}
//
// 	values := make(map[string][]string)
// 	for _, key := range ss {
// 		if v := c.PostForm(key); v != "" {
// 			values[key] = []string{v}
// 		}
// 	}
//
// 	schema.RegisterConverter(game.Phase(0), convertPhase)
// 	schema.RegisterConverter(game.Status(0), convertStatus)
// 	return schema.Decode(g, values)
// }
//
// func convertPhase(value string) reflect.Value {
// 	if v, err := strconv.ParseInt(value, 10, 0); err == nil {
// 		return reflect.ValueOf(game.Phase(v))
// 	}
// 	return reflect.Value{}
// }

// func convertStatus(value string) reflect.Value {
// 	if v, err := strconv.ParseInt(value, 10, 0); err == nil {
// 		return reflect.ValueOf(game.Status(v))
// 	}
// 	return reflect.Value{}
// }

func (g *Game) selectedPlayer() *Player {
	return g.PlayerByID(g.SelectedPlayerID)
}

// BumpedPlayer identifies the player whose theif was bumped to another card due to a played sword.
func (g *Game) BumpedPlayer() *Player {
	return g.PlayerByID(g.BumpedPlayerID)
}

func (g *Game) ID() int64 {
	if g == nil || g.Header == nil || g.Header.Key == nil {
		return -1
	}
	return g.Header.Key.ID
}

func (g *Game) randomTurnOrder() {
	rand.Shuffle(len(g.Players), func(i, j int) { g.Players[i], g.Players[j] = g.Players[j], g.Players[i] })
}