// Package got implements the card game, Guild of Thieves.
package main

import (
	"encoding/json"
	"math/rand"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

// Game stores game state and header information.
type Game struct {
	Key          *datastore.Key `json:"key" datastore:"__key__"`
	EncodedState string         `json:"-" datastore:",noindex"`
	EncodedLog   string         `json:"-" datastore:",noindex"`
	Header
	State `datastore:"-"`
	Log   `datastore:"-"`
}

type Log []map[string]interface{}

// State stores the game state.
type State struct {
	Players             []*Player `json:"players"`
	Log                 GameLog   `json:"-"`
	Grid                Grid      `json:"grid"`
	Jewels              Card      `json:"jewels"`
	SelectedPlayerID    int       `json:"selectedPlayerID"`
	BumpedPlayerID      int       `json:"bumpedPlayerID"`
	SelectedAreaID      areaID    `json:"selectedAreaID"`
	SelectedCardIndex   int       `json:"selectedCardIndex"`
	Stepped             int       `json:"stepped"`
	PlayedCard          *Card     `json:"playedCard"`
	JewelsPlayed        bool      `json:"jewelsPlayed"`
	SelectedThiefAreaID areaID    `json:"selectedThiefAreaID"`
	ClickAreas          []*Area   `json:"clickAreas"`
	Admin               string    `json:"admin"`
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
func (g *History) Start(c *gin.Context) error {
	g.Status = sn.Running
	return g.setupPhase(c)
}

func (g *History) addNewPlayers() {
	for i := range g.UserKeys {
		g.addNewPlayer(i)
	}
}

func (g *History) setupPhase(c *gin.Context) error {
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
// func (g *History) newSetupEntryFor(p *Player) (e *setupEntry) {
// 	e = new(setupEntry)
// 	e.Entry = g.newEntryFor(p)
// 	p.Log = append(p.Log, e)
// 	g.Log = append(g.Log, e)
// 	return
// }
//
// func (e *setupEntry) HTML(g *History) template.HTML {
// 	return restful.HTML("%s received 2 lamps and 1 camel.", g.NameByPID(e.PlayerID))
// }

func (g *History) start(c *gin.Context) error {
	g.Phase = startGame
	// g.newStartEntry()
	return g.placeThieves(c)
}

// type startEntry struct {
// 	*Entry
// }
//
// func (g *History) newStartEntry() *startEntry {
// 	e := new(startEntry)
// 	e.Entry = g.newEntry()
// 	g.Log = append(g.Log, e)
// 	return e
// }
//
// func (e *startEntry) HTML(g *History) template.HTML {
// 	names := make([]string, g.NumPlayers)
// 	for i, p := range g.Players() {
// 		names[i] = g.NameFor(p)
// 	}
// 	return restful.HTML("Good luck %s.  Have fun.", restful.ToSentence(names))
// }

func (h *History) setCurrentPlayer(p *Player) {
	h.CPIDS = nil
	if p != nil {
		h.CPIDS = append(h.CPIDS, p.ID)
	}
}

// PlayerByID returns the player having the provided player id.
func (h *History) PlayerByID(id int) *Player {
	if id <= 0 {
		return nil
	}

	for _, p := range h.Players {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}

// //func (g *History) PlayerBySID(sid string) (p *Player) {
// //	if per := g.Header.PlayerBySID(sid); per != nil {
// //		p = per.(*Player)
// //	}
// //	return
// //}

// PlayerByUserID returns the player having the user id.
func (g *History) PlayerByUserID(id int64) *Player {
	if id <= 0 {
		return nil
	}

	for _, p := range g.Players {
		if p != nil && p.User.ID == id {
			return p
		}
	}
	return nil
}

//func (g *History) PlayerByIndex(index int) (player *Player) {
//	if p := g.PlayererByIndex(index); p != nil {
//		player = p.(*Player)
//	}
//	return
//}

func (g *History) undoTurn(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validateCPorAdmin(c)
	if err != nil {
		return err
	}

	return nil
}

// CurrentPlayer returns the player whose turn it is.
func (h *History) CurrentPlayer() *Player {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	log.Debugf("cpids: %v", h.CPIDS)
	l := len(h.CPIDS)
	if l != 1 {
		return nil
	}
	pid := h.CPIDS[0]
	for _, p := range h.Players {
		if p.ID == pid {
			return p
		}
	}
	return nil
}

// Convenience method for conditionally logging Debug information
// based on package global const debug
//const debug = true
//
//func (g *History) debugf(format string, args ...interface{}) {
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

// func (g *History) adminHeader(c *gin.Context) (string, game.ActionType, error) {
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
// func (g *History) adminUpdateHeader(c *gin.Context, ss sslice) error {
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

func (g *History) selectedPlayer() *Player {
	return g.PlayerByID(g.SelectedPlayerID)
}

// BumpedPlayer identifies the player whose theif was bumped to another card due to a played sword.
func (g *History) BumpedPlayer() *Player {
	return g.PlayerByID(g.BumpedPlayerID)
}

func (g *Game) ID() int64 {
	if g == nil || g.Key == nil {
		return 0
	}

	return g.Key.ID
}

func (g *History) randomTurnOrder() {
	rand.Shuffle(len(g.Players), func(i, j int) { g.Players[i], g.Players[j] = g.Players[j], g.Players[i] })
}

func (g *Game) Load(ps []datastore.Property) error {
	err := datastore.LoadStruct(g, ps)
	if err != nil {
		return err
	}

	var s State
	err = json.Unmarshal([]byte(g.EncodedState), &s)
	if err != nil {
		return err
	}
	g.State = s

	var l Log
	err = json.Unmarshal([]byte(g.EncodedLog), &l)
	if err != nil {
		return err
	}
	g.Log = l
	return nil
}

func (g *Game) Save() ([]datastore.Property, error) {

	encodedState, err := json.Marshal(g.State)
	if err != nil {
		return nil, err
	}
	g.EncodedState = string(encodedState)

	encodedLog, err := json.Marshal(g.Log)
	if err != nil {
		return nil, err
	}
	g.EncodedLog = string(encodedLog)

	t := time.Now()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = t
	}

	g.UpdatedAt = t
	return datastore.SaveStruct(g)
}

func (g *Game) LoadKey(k *datastore.Key) error {
	g.Key = k
	return nil
}

func (g Game) MarshalJSON() ([]byte, error) {
	type JGame Game

	return json.Marshal(struct {
		JGame
		ID           int64   `json:"id"`
		Creator      *User   `json:"creator"`
		Users        []*User `json:"users"`
		LastUpdated  string  `json:"lastUpdated"`
		Public       bool    `json:"public"`
		CreatorEmail omit    `json:"creatorEmail,omitempty"`
		CreatorKey   omit    `json:"creatorKey,omitempty"`
		CreatorName  omit    `json:"creatorName,omitempty"`
		UserEmails   omit    `json:"userEmails,omitempty"`
		UserKeys     omit    `json:"userKeys,omitempty"`
		UserNames    omit    `json:"userNames,omitempty"`
	}{
		JGame:       JGame(g),
		ID:          g.ID(),
		Creator:     toUser(g.CreatorKey, g.CreatorName, g.CreatorEmail),
		Users:       toUsers(g.UserKeys, g.UserNames, g.UserEmails),
		LastUpdated: sn.LastUpdated(g.UpdatedAt),
		Public:      g.Password == "",
	})
}

func rootKey(id int64) *datastore.Key {
	return datastore.IDKey(rootKind, id, nil)
}
