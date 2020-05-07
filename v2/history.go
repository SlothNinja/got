package main

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

type History Game

func newHistory(id, rev int64) *History {
	h := new(History)
	h.Key = newHistoryKey(id, rev)
	h.Type = sn.GOT
	return h
}

func newHistoryKey(id, rev int64) *datastore.Key {
	return datastore.NameKey(historyKind, fmt.Sprintf("%d-%d", id, rev), rootKey(id))
}

func (h *History) ID() int64 {
	if h == nil || h.Key == nil || h.Key.Parent == nil {
		return 0
	}
	return h.Key.Parent.ID
}

func (client Client) getHistory(c *gin.Context, inc ...int64) (*History, error) {
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

	h := newHistory(id, undo.Current)
	err = client.DS.Get(c, h.Key, h)
	if err != nil {
		return nil, err
	}
	h.Undo = undo
	return h, nil
}

func (g *History) Load(ps []datastore.Property) error {
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

func (g *History) Save() ([]datastore.Property, error) {

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

func (g *History) LoadKey(k *datastore.Key) error {
	g.Key = k
	return nil
}

func (g History) MarshalJSON() ([]byte, error) {
	type JHistory History

	return json.Marshal(struct {
		JHistory
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
		JHistory:    JHistory(g),
		ID:          g.ID(),
		Creator:     toUser(g.CreatorKey, g.CreatorName, g.CreatorEmail),
		Users:       toUsers(g.UserKeys, g.UserNames, g.UserEmails),
		LastUpdated: sn.LastUpdated(g.UpdatedAt),
		Public:      g.Password == "",
	})
}
