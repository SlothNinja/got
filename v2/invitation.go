package main

import (
	"encoding/json"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Pallinder/go-randomdata"
	"github.com/SlothNinja/sn/v2"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

// Invitation provides a game invitation
type Invitation struct {
	Key *datastore.Key `json:"key" datastore:"__key__"`
	Header
}

func (inv *Invitation) ID() int64 {
	if inv == nil || inv.Key == nil {
		return 0
	}
	return inv.Key.ID
}

func newInvitation(id int64) *Invitation {
	return &Invitation{Key: newInvitationKey(id)}
}

func newInvitationKey(id int64) *datastore.Key {
	return datastore.IDKey(invitationKind, id, rootKey(id))
}

func (inv *Invitation) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(inv, ps)
}

func (inv *Invitation) Save() ([]datastore.Property, error) {
	t := time.Now()
	if inv.CreatedAt.IsZero() {
		inv.CreatedAt = t
	}
	inv.UpdatedAt = t
	return datastore.SaveStruct(inv)
}

func (inv *Invitation) LoadKey(k *datastore.Key) error {
	inv.Key = k
	return nil
}

func defaultInvitation() *Invitation {
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

type omit *struct{}

func (inv Invitation) MarshalJSON() ([]byte, error) {
	type JInvitation Invitation

	return json.Marshal(struct {
		JInvitation
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
		JInvitation: JInvitation(inv),
		ID:          inv.ID(),
		Creator:     toUser(inv.CreatorKey, inv.CreatorName, inv.CreatorEmail),
		Users:       toUsers(inv.UserKeys, inv.UserNames, inv.UserEmails),
		LastUpdated: sn.LastUpdated(inv.UpdatedAt),
		Public:      inv.Password == "",
	})
}

type User struct {
	*user.User
	ID        int64 `json:"id"`
	LCName    omit  `json:"lcname,omitempty"`
	Joined    omit  `json:"joined,omitempty"`
	CreatedAt omit  `json:"createdat,omitempty"`
	UpdatedAt omit  `json:"updatedat,omitempty"`
	Admin     omit  `json:"admin,omitempty"`
}

func toUser(k *datastore.Key, name, email string) *User {
	var id int64 = -1
	if k != nil {
		id = k.ID
	}
	u := &User{User: user.New(id)}
	u.ID = id
	u.Name = name
	u.Email = email
	return u
}

func toUsers(ks []*datastore.Key, names, emails []string) []*User {
	us := make([]*User, len(ks))
	for i := range ks {
		us[i] = toUser(ks[i], names[i], emails[i])
	}
	return us
}
