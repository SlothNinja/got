package main

import (
	"strconv"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

// func newHeaderEntity(g *Game) *headerEntity {
// 	return &headerEntity{Header: g.Header, Key: newHeaderKey(g.ID())}
// }
//
// func newHeaderKey(id int64) *datastore.Key {
// 	return datastore.IDKey(headerKind, id, nil)
// }
//
// type headerEntity struct {
// 	Key *datastore.Key `datastore:"__key__" json:"-"`
// 	Header
// }
//
// func (e *headerEntity) Accept(c *gin.Context, cu *user.User) (bool, error) {
// 	return e.Header.Header.Accept(c, cu)
// }
//
// func (e *headerEntity) Drop(cu *user.User) error {
// 	return e.Header.Header.Drop(cu)
// }
//
// func (e *headerEntity) AddUser(cu *user.User) {
// 	e.Header.Header.AddUser(cu)
// }
//
// func (e headerEntity) LastUpdated() string {
// 	return restful.LastUpdated(e.UpdatedAt)
// }
//
// func (e headerEntity) LastUpdate() time.Time {
// 	return e.UpdatedAt
// }
//
// func (e headerEntity) ID() int64 {
// 	return e.Key.ID
// }
//
// func (e headerEntity) MarshalJSON() ([]byte, error) {
// 	status := "Public"
// 	if e.Password != "" {
// 		status = "Private"
// 	}
//
// 	type JEntity headerEntity
// 	return json.Marshal(struct {
// 		JEntity
// 		ID          int64  `json:"id"`
// 		Public      string `json:"public"`
// 		LastUpdated string `json:"lastUpdated"`
// 	}{
// 		JEntity:     JEntity(e),
// 		ID:          e.Key.ID,
// 		Public:      status,
// 		LastUpdated: e.LastUpdated(),
// 	})
// }

// Header provides game/invitation header data
type Header struct {
	Key             *datastore.Key `datastore:"__key__"`
	TwoThiefVariant bool           `json:"twoThief"`
	sn.Header
}

func newHeader(id int64) *Header {
	return &Header{Key: newHeaderKey(id)}
}

func newHeaderKey(id int64) *datastore.Key {
	return datastore.IDKey(headerKind, id, nil)
}

func getHID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("hid"), 10, 64)
}
