package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
)

type Game struct {
	Key          *datastore.Key `datastore:"__key__"`
	EncodedState string         `datastore:",noindex"`
	EncodedLog   string         `datastore:",noindex"`
	Header
	glog
	state
}

func newGame(id, rev int64) *Game {
	return &Game{Key: newGameKey(id, rev)}
}

func newGameKey(id, rev int64) *datastore.Key {
	return datastore.NameKey(gameKind, fmt.Sprintf("%d-%d", id, rev), rootKey(id))
}

func (g *Game) id() int64 {
	if g.Key == nil || g.Key.Parent == nil {
		return 0
	}
	return g.Key.Parent.ID
}

func (g *Game) rev() int64 {
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

func (g *Game) Load(ps []datastore.Property) error {
	err := datastore.LoadStruct(g, ps)
	if err != nil {
		return err
	}

	var s state
	err = json.Unmarshal([]byte(g.EncodedState), &s)
	if err != nil {
		return err
	}
	g.state = s

	var l glog
	err = json.Unmarshal([]byte(g.EncodedLog), &l)
	if err != nil {
		return err
	}
	g.glog = l
	return nil
}

func (g *Game) Save() ([]datastore.Property, error) {

	encodedState, err := json.Marshal(g.state)
	if err != nil {
		return nil, err
	}
	g.EncodedState = string(encodedState)

	encodedLog, err := json.Marshal(g.glog)
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

func (g *Game) MarshalJSON() ([]byte, error) {
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

	s, err := json.Marshal(g.state)
	if err != nil {
		return nil, err
	}

	var state map[string]interface{}
	err = json.Unmarshal(s, &state)
	if err != nil {
		return nil, err
	}

	data["key"] = g.Key
	data["id"] = g.id()
	data["log"] = g.glog
	data["rev"] = g.rev()

	for k, v := range state {
		data[k] = v
	}

	return json.Marshal(data)
}

// func (cl *client) getUsers() error {
// 	g.Users = make([]*user.User, len(g.UserIDS))
// 	haveCreator := false
//
// 	var err error
// 	for i, uid := range g.UserIDS {
// 		g.Users[i], err = cl.User.Get(cl.ctx, uid)
// 		if err != nil {
// 			g.Users = nil
// 			return err
// 		}
// 		if g.CreatorID == uid {
// 			g.Creator = g.Users[i]
// 			haveCreator = true
// 		}
// 	}
//
// 	if haveCreator {
// 		return nil
// 	}
//
// 	g.Creator, err = cl.User.Get(cl.ctx, g.CreatorID)
// 	return err
// }
