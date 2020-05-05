package main

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/sn/v2"
)

// Header provides game/invitation header data
type Header struct {
	TwoThiefVariant bool  `json:"twoThief"`
	Phase           Phase `json:"phase"`
	sn.Header
}

type GHeader struct {
	Key *datastore.Key `datastore:"__key__"`
	Header
}

func newGHeader(id int64) *GHeader {
	return &GHeader{Key: newGHeaderKey(id)}
}

func newGHeaderKey(id int64) *datastore.Key {
	return datastore.IDKey(gheaderKind, id, rootKey(id))
}

func (gh *GHeader) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(gh, ps)
}

func (gh *GHeader) Save() ([]datastore.Property, error) {
	t := time.Now()
	if gh.CreatedAt.IsZero() {
		gh.CreatedAt = t
	}
	gh.UpdatedAt = t
	return datastore.SaveStruct(gh)
}

func (gh *GHeader) LoadKey(k *datastore.Key) error {
	gh.Key = k
	return nil
}

func (gh GHeader) MarshalJSON() ([]byte, error) {
	type JGHeader GHeader

	return json.Marshal(struct {
		JGHeader
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
		JGHeader:    JGHeader(gh),
		ID:          gh.Key.ID,
		Creator:     toUser(gh.CreatorKey, gh.CreatorName, gh.CreatorEmail),
		Users:       toUsers(gh.UserKeys, gh.UserNames, gh.UserEmails),
		LastUpdated: sn.LastUpdated(gh.UpdatedAt),
		Public:      gh.Password == "",
	})
}
