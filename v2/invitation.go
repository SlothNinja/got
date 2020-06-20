package main

import (
	"encoding/json"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Pallinder/go-randomdata"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

type invitation struct {
	Key *datastore.Key
	Header
}

func (inv invitation) MarshalJSON() ([]byte, error) {
	h, err := json.Marshal(inv.Header)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(h, &data)
	if err != nil {
		return nil, err
	}

	data["key"] = inv.Key
	data["id"] = inv.ID()
	data["lastUpdated"] = sn.LastUpdated(inv.UpdatedAt)
	data["public"] = len(inv.Password) == 0

	return json.Marshal(data)
}

func (inv *invitation) ID() int64 {
	if inv == nil || inv.Key == nil {
		return 0
	}
	return inv.Key.ID
}

func newInvitation(id int64) *invitation {
	return &invitation{Key: newInvitationKey(id)}
}

func newInvitationKey(id int64) *datastore.Key {
	return datastore.IDKey(invitationKind, id, rootKey(id))
}

func (inv *invitation) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(inv, ps)
}

func (inv *invitation) Save() ([]datastore.Property, error) {
	t := time.Now()
	if inv.CreatedAt.IsZero() {
		inv.CreatedAt = t
	}
	inv.UpdatedAt = t
	return datastore.SaveStruct(inv)
}

func (inv *invitation) LoadKey(k *datastore.Key) error {
	inv.Key = k
	return nil
}

func defaultInvitation() *invitation {
	inv := newInvitation(0)

	// Default Values
	inv.Title = randomdata.SillyName()
	inv.NumPlayers = 2
	inv.TwoThiefVariant = false
	return inv
}

func getID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param(idParam), 10, 64)
}

// type omit *struct{}
//
// type User struct {
// 	*user.User
// 	ID        int64 `json:"id"`
// 	LCName    omit  `json:"lcname,omitempty"`
// 	Joined    omit  `json:"joined,omitempty"`
// 	CreatedAt omit  `json:"createdat,omitempty"`
// 	UpdatedAt omit  `json:"updatedat,omitempty"`
// 	Admin     omit  `json:"admin,omitempty"`
// }
//
// func toUser(k *datastore.Key, name, email string) *User {
// 	var id int64 = -1
// 	if k != nil {
// 		id = k.ID
// 	}
// 	u := &User{User: user.New(id)}
// 	u.ID = id
// 	u.Name = name
// 	u.Email = email
// 	return u
// }
//
// func toUsers(ks []*datastore.Key, names, emails []string) []*User {
// 	us := make([]*User, len(ks))
// 	for i := range ks {
// 		us[i] = toUser(ks[i], names[i], emails[i])
// 	}
// 	return us
// }
