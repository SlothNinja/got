package main

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/sn"
)

// Header provides game/invitation header data
type Header struct {
	TwoThiefVariant bool
	Phase           phase
	game.Header
}

// MarshalJSON implements json.Marshaler interface
func (h Header) MarshalJSON() ([]byte, error) {
	snh, err := json.Marshal(h.Header)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(snh, &data)
	if err != nil {
		return nil, err
	}

	data["twoThief"] = h.TwoThiefVariant
	data["phase"] = h.Phase

	return json.Marshal(data)
}

// GHeader stores game headers with associate game data.
type GHeader struct {
	Key *datastore.Key `datastore:"__key__"`
	Header
}

func (gh GHeader) id() int64 {
	if gh.Key == nil {
		return 0
	}
	return gh.Key.ID
}

// MarshalJSON implements json.Marshaler interface
func (gh GHeader) MarshalJSON() ([]byte, error) {
	h, err := json.Marshal(gh.Header)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(h, &data)
	if err != nil {
		return nil, err
	}

	data["key"] = gh.Key
	data["id"] = gh.id()
	data["lastUpdated"] = sn.LastUpdated(gh.UpdatedAt)
	data["public"] = len(gh.Password) == 0

	return json.Marshal(data)
}

func newGHeader(id int64) *GHeader {
	return &GHeader{Key: newGHeaderKey(id)}
}

func newGHeaderKey(id int64) *datastore.Key {
	return datastore.IDKey(headerKind, id, rootKey(id))
}

// Load implements datastore.PropertyLoadSaver interface
func (gh *GHeader) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(gh, ps)
}

// Save implements datastore.PropertyLoadSaver interface
func (gh *GHeader) Save() ([]datastore.Property, error) {
	t := time.Now()
	if gh.CreatedAt.IsZero() {
		gh.CreatedAt = t
	}
	gh.UpdatedAt = t
	return datastore.SaveStruct(gh)
}

// LoadKey implements datastore.LoadKey interface
func (gh *GHeader) LoadKey(k *datastore.Key) error {
	gh.Key = k
	return nil
}
