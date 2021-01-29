// Package got implements the card game, Guild of Thieves.
package got

import (
	"encoding/gob"
	"html/template"
	"reflect"
	"strconv"

	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/schema"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register(new(setupEntry))
	gob.Register(new(startEntry))
}

// Register assigns a game type and routes.
func (client *Client) register(t gtype.Type) *Client {
	gob.Register(new(Game))
	game.Register(t, newGamer, phaseNames, nil)
	return client.addRoutes(t.Prefix())
}

//var ErrMustBeGame = errors.New("Resource must have type *Game.")

const noPID = game.NoPlayerID

// Game stores game state and header information.
type Game struct {
	*game.Header
	*State
}

// State stores the game state.
type State struct {
	Playerers          game.Playerers
	Log                GameLog
	Grid               grid
	Jewels             Card
	TwoThiefVariant    bool `form:"two-thief-variant"`
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

// GetPlayerers implements the GetPlayerers interfaces of the sn/games package.
// Generally used to support common player manipulation functions of sn/games package.
func (g *Game) GetPlayerers() game.Playerers {
	return g.Playerers
}

// Players returns a slice of player structs that store various information about each player.
func (g *Game) Players() []*Player {
	pers := g.GetPlayerers()
	length := len(pers)
	if length == 0 {
		return nil
	}
	ps := make([]*Player, length)
	for i, p := range pers {
		ps[i] = p.(*Player)
	}
	return ps
}

func (g *Game) setPlayers(ps Players) {
	length := len(ps)
	if length > 0 {
		pers := make(game.Playerers, length)
		for i, p := range ps {
			pers[i] = p
		}
		g.Playerers = pers
	}
}

// Start begins a Guild of Thieves game.
func (client *Client) Start() {
	client.Game.Status = game.Running
	client.setupPhase()
}

func (g *Game) addNewPlayers() {
	for _, u := range g.Users {
		g.addNewPlayer(u)
	}
}

func (client *Client) setupPhase() {
	client.Game.Turn = 0
	client.Game.Phase = setup
	client.Game.addNewPlayers()
	client.Game.RandomTurnOrder()
	client.Game.createGrid()
	for _, p := range client.Game.Players() {
		client.newSetupEntryFor(p)
	}
	cp := client.previousPlayer(client.Game.Players()[0])
	client.setCurrentPlayers(cp)
	client.beginningOfPhaseReset()
	client.start()
}

type setupEntry struct {
	*Entry
}

func (client *Client) newSetupEntryFor(p *Player) *setupEntry {
	e := new(setupEntry)
	e.Entry = client.newEntryFor(p)
	p.Log = append(p.Log, e)
	client.Game.Log = append(client.Game.Log, e)
	return e
}

func (e *setupEntry) HTML(g *Game) template.HTML {
	return restful.HTML("%s received 2 lamps and 1 camel.", g.NameByPID(e.PlayerID))
}

func (client *Client) start() {
	client.Game.Phase = startGame
	client.newStartEntry()
	client.placeThieves()
}

type startEntry struct {
	*Entry
}

func (client *Client) newStartEntry() *startEntry {
	e := new(startEntry)
	e.Entry = client.newEntry()
	client.Game.Log = append(client.Game.Log, e)
	return e
}

func (e *startEntry) HTML(g *Game) template.HTML {
	names := make([]string, g.NumPlayers)
	for i, p := range g.Players() {
		names[i] = g.NameFor(p)
	}
	return restful.HTML("Good luck %s.  Have fun.", restful.ToSentence(names))
}

func (client *Client) setCurrentPlayers(ps ...*Player) {
	var pers game.Playerers

	switch length := len(ps); {
	case length == 0:
		pers = nil
	case length == 1:
		pers = game.Playerers{ps[0]}
	default:
		pers = make(game.Playerers, length)
		for i, player := range ps {
			pers[i] = player
		}
	}
	client.Game.SetCurrentPlayerers(pers...)
}

// PlayerByID returns the player having the provided player id.
func (g *Game) PlayerByID(id int) (p *Player) {
	if per := g.PlayererByID(id); per != nil {
		p = per.(*Player)
	}
	return
}

//func (g *Game) PlayerBySID(sid string) (p *Player) {
//	if per := g.Header.PlayerBySID(sid); per != nil {
//		p = per.(*Player)
//	}
//	return
//}

// PlayerByUserID returns the player having the user id.
func (g *Game) PlayerByUserID(id int64) (player *Player) {
	if p := g.PlayererByUserID(id); p != nil {
		player = p.(*Player)
	}
	return
}

//func (g *Game) PlayerByIndex(index int) (player *Player) {
//	if p := g.PlayererByIndex(index); p != nil {
//		player = p.(*Player)
//	}
//	return
//}

// func (client *Client) undoTurn(c *gin.Context) {
// 	client.Log.Debugf(msgEnter)
// 	defer client.Log.Debugf(msgExit)
//
// 	path := showPath(client.Prefix, c.Param("hid"))
//
// 	if !client.Game.IsCurrentPlayer(client.CUser) {
// 		restful.AddErrorf(c, "%v", sn.NewVError("only the current player may perform this action"))
// 		c.Redirect(http.StatusSeeOther, path)
// 	}
//
// 	cp := client.Game.CurrentPlayer()
// 	if cp != nil {
// 		restful.AddNoticef(c, "%s undid turn.", client.Game.NameFor(cp))
// 	}
//
// 	client.Cache.Delete(clclient.CUser)
//
// 	c.Redirect(http.StatusSeeOther, path)
// }

// CurrentPlayer returns the player whose turn it is.
func (g *Game) CurrentPlayer() *Player {
	p := g.CurrentPlayerer()
	if p != nil {
		return p.(*Player)
	}
	return nil
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

func (g *Game) adminHeader(c *gin.Context, cu *user.User) (string, game.ActionType, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = selectThief
	g.CurrentPlayer().PerformedAction = false
	g.SelectedAreaF = nil
	g.PlayedCard = newCard(lamp, true)

	// if err := g.adminUpdateHeader(c, headerValues); err != nil {
	// 	return "got/flash_notice", game.None, err
	// }

	return "", game.Cache, nil
}

func (client *Client) adminUpdateHeader(ss sslice) error {
	err := client.validateAdminAction()
	if err != nil {
		return err
	}

	values := make(map[string][]string)
	for _, key := range ss {
		if v := client.Context.PostForm(key); v != "" {
			values[key] = []string{v}
		}
	}

	schema.RegisterConverter(game.Phase(0), convertPhase)
	schema.RegisterConverter(game.Status(0), convertStatus)
	return schema.Decode(client.Game, values)
}

func convertPhase(value string) reflect.Value {
	if v, err := strconv.ParseInt(value, 10, 0); err == nil {
		return reflect.ValueOf(game.Phase(v))
	}
	return reflect.Value{}
}

func convertStatus(value string) reflect.Value {
	if v, err := strconv.ParseInt(value, 10, 0); err == nil {
		return reflect.ValueOf(game.Status(v))
	}
	return reflect.Value{}
}

func (g *Game) selectedPlayer() *Player {
	return g.PlayerByID(g.SelectedPlayerID)
}

// BumpedPlayer identifies the player whose theif was bumped to another card due to a played sword.
func (g *Game) BumpedPlayer() *Player {
	return g.PlayerByID(g.BumpedPlayerID)
}
