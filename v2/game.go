package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

type Game struct {
	Key          *datastore.Key `datastore:"__key__"`
	EncodedState string         `datastore:",noindex"`
	EncodedLog   string         `datastore:",noindex"`
	Header
	State `datastore:"-"`
	Log   `datastore:"-"`
}

func newGame(id, rev int64) *Game {
	g := new(Game)
	g.Key = newGameKey(id, rev)
	g.Type = sn.GOT
	return g
}

func newGameKey(id, rev int64) *datastore.Key {
	return datastore.NameKey(gameKind, fmt.Sprintf("%d-%d", id, rev), rootKey(id))
}

func (g Game) ID() int64 {
	if g.Key == nil || g.Key.Parent == nil {
		return 0
	}
	return g.Key.Parent.ID
}

func (g Game) Rev() int64 {
	if g.Key == nil {
		return 0
	}
	s := strings.Split(g.Key.Name, "-")
	if len(s) != 2 {
		return g.Undo.Current
	}
	rev, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		log.Warningf(err.Error())
		return 0
	}
	return rev
}

func (client Client) getGame(c *gin.Context, inc ...int64) (*Game, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	id, err := getID(c)
	if err != nil {
		return nil, err
	}

	undo, err := getStack(c)
	if err != nil {
		return nil, err
	}

	if len(inc) == 1 {
		undo.Current += inc[0]
	}

	g := newGame(id, undo.Current)
	err = client.DS.Get(c, g.Key, g)
	if err != nil {
		return nil, err
	}
	g.Undo = undo
	return g, nil
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
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	hh, err := json.Marshal(g.Header)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(hh, &data)
	if err != nil {
		return nil, err
	}

	s, err := json.Marshal(g.State)
	if err != nil {
		return nil, err
	}

	var state map[string]interface{}
	err = json.Unmarshal(s, &state)
	if err != nil {
		return nil, err
	}

	data["key"] = g.Key
	data["id"] = g.ID()
	data["log"] = g.Log
	data["rev"] = g.Rev()

	for k, v := range state {
		data[k] = v
	}

	return json.Marshal(data)
}
