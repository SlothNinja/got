package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (cl client) show(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := cl.getGCommited(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	g.updateClickablesFor(c, g.currentPlayer(), g.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (cl client) undo(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cl.undoOperations(c, (*sn.Stack).Undo)
}

func (cl client) redo(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cl.undoOperations(c, (*sn.Stack).Redo)
}

func (cl client) reset(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	gcommitted, err := cl.getGCommited(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	k := newGameKey(gcommitted.id(), gcommitted.Undo.Current)
	g := gcommitted.game
	_, err = cl.DS.Put(c, k, &g)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	g.updateClickablesFor(c, g.currentPlayer(), g.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (cl client) undoOperations(c *gin.Context, action func(*sn.Stack) bool) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	undo, err := getStack(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	id, err := getID(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	action(&undo)
	g := newGame(id, undo.Current)
	err = cl.DS.Get(c, g.Key, g)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if undo.Committed != g.Undo.Committed {
		sn.JErr(c, fmt.Errorf("invalid game state"))
		return
	}

	g.Undo = undo
	g.updateClickablesFor(c, g.currentPlayer(), g.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (cl client) newInvitation(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cu, err := user.FromSession(c)
	if err != nil || cu == nil {
		sn.JErr(c, sn.ErrUserNotFound)
		return
	}

	inv := defaultInvitation()

	c.JSON(http.StatusOK, gin.H{"invitation": inv})
}

func (cl client) create(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cu, err := user.FromSession(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	inv := newInvitation(0)
	err = inv.fromForm(c, cu)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	ks, err := cl.DS.AllocateIDs(c, []*datastore.Key{rootKey(0)})
	if err != nil {
		sn.JErr(c, err)
		return
	}
	inv.Key = newInvitationKey(ks[0].ID)

	_, err = cl.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		m := sn.NewMLog(inv.Key.ID)
		ks := []*datastore.Key{inv.Key, m.Key}
		es := []interface{}{inv, m}

		_, err := tx.PutMulti(ks, es)
		return err
	})
	if err != nil {
		sn.JErr(c, err)
		return
	}

	inv2 := defaultInvitation()
	c.JSON(http.StatusOK, gin.H{
		"invitation": inv2,
		"message":    fmt.Sprintf("%s created game %q", cu.Name, inv.Title),
	})
}

func (inv *invitation) fromForm(c *gin.Context, cu *user.User) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	obj := struct {
		Title           string `form:"title"`
		NumPlayers      int    `form:"num-players" binding:"min=0,max=5"`
		Password        string `form:"password"`
		TwoThiefVariant bool   `form:"two-thief-variant"`
	}{}

	err := c.ShouldBind(&obj)
	if err != nil {
		return err
	}

	inv.Title = cu.Name + "'s Game"
	if obj.Title != "" {
		inv.Title = obj.Title
	}

	inv.NumPlayers = 4
	if obj.NumPlayers >= 1 && obj.NumPlayers <= 5 {
		inv.NumPlayers = obj.NumPlayers
	}

	if len(obj.Password) > 0 {
		hashed, err := bcrypt.GenerateFromPassword([]byte(obj.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		inv.Password = hashed
	}
	inv.AddCreator(cu)
	inv.TwoThiefVariant = obj.TwoThiefVariant
	inv.AddUser(cu)
	inv.Status = sn.Recruiting
	inv.Type = sn.GOT
	return nil
}

func (cl client) accept(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	inv, err := cl.getInvitation(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cu, err := user.FromSession(c)
	if err != nil || cu == nil {
		sn.JErr(c, sn.ErrUserNotFound)
		return
	}

	obj := struct {
		Password string `json:"password"`
	}{}

	err = c.ShouldBind(&obj)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	start, err := inv.Accept(cu, []byte(obj.Password))
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if !start {
		_, err = cl.DS.Put(c, inv.Key, inv)
		if err != nil {
			sn.JErr(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"invitation": inv,
			"message":    fmt.Sprintf("%s joined game: %d", cu.Name, inv.Key.ID),
		})
		return
	}

	g := newGame(inv.Key.ID, 0)
	g.Header = inv.Header
	g.start()

	_, err = cl.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		err = tx.Delete(inv.Key)
		if err != nil {
			return err
		}

		g.StartedAt = time.Now()
		_, err = tx.PutMulti(g.save())
		return err
	})
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cp := g.currentPlayer()
	err = g.SendTurnNotificationsTo(c, cp)
	if err != nil {
		log.Warningf(err.Error())
	}

	inv.Header = g.Header
	c.JSON(http.StatusOK, gin.H{
		"invitation": inv,
		"message": fmt.Sprintf(
			`<div>Game: %d has started.</div>
			<div></div>
			<div><strong>%s</strong> is start player.</div>`,
			inv.ID(), cp.User.Name),
	})
}

type detail struct {
	ID  int64 `json:"id"`
	GLO int   `json:"glo"`
}

func (cl client) details(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	inv, err := cl.getInvitation(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cu, err := user.FromSession(c)
	if err != nil || cu == nil {
		sn.JErr(c, sn.ErrUserNotFound)
		return
	}

	ks := make([]*datastore.Key, len(inv.UserKeys))
	copy(ks, inv.UserKeys)

	hasKey := false
	for _, k := range inv.UserKeys {
		if k.Equal(cu.Key) {
			hasKey = true
			break
		}
	}

	if !hasKey {
		ks = append(ks, cu.Key)
	}

	ratings, err := cl.SN.GetMulti(c, ks, inv.Type)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	details := make([]detail, len(ratings))
	for i, rating := range ratings {
		details[i] = detail{ID: rating.Key.Parent.ID, GLO: rating.Rank().GLO()}
	}

	c.JSON(http.StatusOK, gin.H{"details": details})
}

func (g *game) cache() (*datastore.Key, interface{}) {
	return newGameKey(g.id(), g.Undo.Current), g
}

func (g *game) save() ([]*datastore.Key, []interface{}) {
	gh := newGHeader(g.id())
	gh.Header = g.Header

	ks := []*datastore.Key{newGCommittedKey(g.id()), newGameKey(g.id(), g.Undo.Current), gh.Key}
	es := []interface{}{g, g, gh}
	return ks, es
}

func (cl client) drop(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	inv, err := cl.getInvitation(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cu, err := user.FromSession(c)
	if err != nil || cu == nil {
		sn.JErr(c, sn.ErrUserNotFound)
		return
	}

	err = inv.Drop(cu)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if len(inv.UserKeys) == 0 {
		inv.Status = sn.Aborted
	}

	_, err = cl.DS.Put(c, inv.Key, inv)
	if err != nil {
		sn.JErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"invitation": inv,
		"message":    fmt.Sprintf("%s dropped from game invitation: %d", cu.Name, inv.Key.ID),
	})
}

func (cl client) getInvitation(c *gin.Context) (*invitation, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	id, err := getID(c)
	if err != nil {
		return nil, err
	}

	inv := newInvitation(id)
	err = cl.DS.Get(c, inv.Key, inv)
	return inv, err
}

func (cl client) getGCommited(c *gin.Context) (*gcommitted, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	id, err := getID(c)
	if err != nil {
		return nil, err
	}

	g := newGCommited(id)
	err = cl.DS.Get(c, g.Key, g)
	return g, err
}

func getStack(c *gin.Context) (sn.Stack, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	var emptyStack sn.Stack

	bodyBytes, err := copyBody(c)
	if err != nil {
		return emptyStack, err
	}

	obj := struct {
		sn.Stack `json:"undo"`
	}{}
	err = c.ShouldBind(&obj)
	log.Debugf("err: %v\nobj: %#v", err, obj)
	if err != nil {
		return sn.Stack{}, err
	}

	// Restore the io.ReadCloser to its original state
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return obj.Stack, nil
}

// copyBody returns a copy of c.Request.Body and resets to permit further reading of c.Request.Body
func copyBody(c *gin.Context) ([]byte, error) {
	// Read the content
	if c.Request.Body == nil {
		return nil, fmt.Errorf("request missing body")
	}

	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	cpBytes := make([]byte, len(bodyBytes))
	copy(cpBytes, bodyBytes)

	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return cpBytes, nil
}

func (cl client) invitationsIndex(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cu, err := user.FromSession(c)
	if err != nil {
		sn.JErr(c, err)
	}

	if cu == nil {
		sn.JErr(c, sn.ErrUserNotFound)
		return
	}

	q := datastore.
		NewQuery(invitationKind).
		Filter("Status=", int(sn.Recruiting)).
		Order("-UpdatedAt")

	var es []*invitation
	_, err = cl.DS.GetAll(c, q, &es)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"invitations": es, "cu": cu})
}

func (cl client) gamesIndex(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cu, err := user.FromSession(c)
	if err != nil {
		sn.JErr(c, err)
	}

	if cu == nil {
		sn.JErr(c, sn.ErrUserNotFound)
		return
	}

	status := sn.ToStatus[c.Param("status")]
	q := datastore.
		NewQuery(headerKind).
		Filter("Status=", int(status)).
		Order("-UpdatedAt")

	var es []*GHeader
	_, err = cl.DS.GetAll(c, q, &es)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"gheaders": es, "cu": cu})
}

func (cl client) current(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	u, err := user.FromSession(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cu": sn.ToUser(u.Key, u.Name, u.EmailHash)})
}

// func jerr(c *gin.Context, err error) {
// 	if errors.Is(err, sn.ErrValidation) {
// 		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
// 		return
// 	}
// 	log.Debugf(err.Error())
// 	c.JSON(http.StatusOK, gin.H{"message": sn.ErrUnexpected.Error()})
// }
