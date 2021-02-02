package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/sn"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/undo"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (cl *client) show(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	var err error
	err = cl.getGCommited()
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cl.updateClickablesFor(cl.currentPlayer(), cl.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": cl.g})
}

func (cl *client) undo(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c
	cl.undoOperations((*undo.Stack).Undo)
}

func (cl *client) redo(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c
	cl.undoOperations((*undo.Stack).Redo)
}

func (cl *client) reset(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	err := cl.getGCommited()
	if err != nil {
		sn.JErr(c, err)
		return
	}

	k := newGameKey(cl.gc.id(), cl.gc.Undo.Current)
	cl.g = cl.gc.Game
	_, err = cl.DS.Put(c, k, cl.g)
	if err != nil {
		cl.jerr(err)
		return
	}

	cl.updateClickablesFor(cl.currentPlayer(), cl.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": cl.g})
}

func (cl *client) undoOperations(action func(*undo.Stack) bool) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	undo, err := cl.getStack()
	if err != nil {
		cl.jerr(err)
		return
	}

	id, err := cl.getID()
	if err != nil {
		cl.jerr(err)
		return
	}

	action(undo)
	cl.g = newGame(id, undo.Current)
	err = cl.DS.Get(cl.ctx, cl.g.Key, cl.g)
	if err != nil {
		cl.jerr(err)
		return
	}

	if undo.Committed != cl.g.Undo.Committed {
		cl.jerr(fmt.Errorf("invalid game state"))
		return
	}

	cl.g.Undo = undo
	cl.updateClickablesFor(cl.currentPlayer(), cl.selectedThiefArea())
	cl.ctx.JSON(http.StatusOK, gin.H{"game": cl.g})
}

func (cl *client) newInvitationHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	cl.CUser()
	if cl.cu == nil {
		cl.jerr(user.ErrNotFound)
		return
	}

	inv := defaultInvitation()

	c.JSON(http.StatusOK, gin.H{"invitation": inv})
}

func (cl *client) createHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	cl.CUser()
	if cl.cu == nil {
		cl.jerr(user.ErrNotFound)
		return
	}

	inv := newInvitation(0)
	err := inv.fromForm(c, cl.cu)
	if err != nil {
		cl.jerr(err)
		return
	}

	ks, err := cl.DS.AllocateIDs(c, []*datastore.Key{rootKey(0)})
	if err != nil {
		cl.jerr(err)
		return
	}
	inv.Key = newInvitationKey(ks[0].ID)

	_, err = cl.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		m := mlog.New(inv.Key.ID)
		ks := []*datastore.Key{inv.Key, m.Key}
		es := []interface{}{inv, m}

		_, err := tx.PutMulti(ks, es)
		return err
	})
	if err != nil {
		cl.jerr(err)
		return
	}

	inv2 := defaultInvitation()
	c.JSON(http.StatusOK, gin.H{
		"invitation": inv2,
		"message":    fmt.Sprintf("%s created game %q", cl.cu.Name, inv.Title),
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
		inv.PasswordHash = hashed
	}
	inv.AddCreator(cu)
	inv.TwoThiefVariant = obj.TwoThiefVariant
	inv.AddUser(cu)
	inv.Status = game.Recruiting
	inv.Type = gtype.GOT
	return nil
}

func (cl *client) accept(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	inv, err := cl.getInvitation()
	if err != nil {
		cl.jerr(err)
		return
	}

	cu := cl.CUser()
	if cu == nil {
		cl.jerr(user.ErrNotFound)
		return
	}

	obj := struct {
		Password string `json:"password"`
	}{}

	err = c.ShouldBind(&obj)
	if err != nil {
		cl.jerr(err)
		return
	}

	start, err := inv.AcceptWith(cu, []byte(obj.Password))
	if err != nil {
		cl.jerr(err)
		return
	}

	if !start {
		_, err = cl.DS.Put(c, inv.Key, inv)
		if err != nil {
			cl.jerr(err)
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
	cl.start()

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
		cl.jerr(err)
		return
	}

	err = cl.sendTurnNotificationsTo(cl.cp)
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
			inv.ID(), cl.nameFor(cl.cp)),
	})
}

type detail struct {
	ID  int64 `json:"id"`
	GLO int   `json:"glo"`
}

func (cl *client) details(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	inv, err := cl.getInvitation()
	if err != nil {
		cl.jerr(err)
		return
	}

	cu := cl.CUser()
	if cu == nil {
		cl.jerr(user.ErrNotFound)
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

	ratings, err := cl.Rating.GetMulti(c, ks, gtype.Type(inv.Type))
	if err != nil {
		cl.jerr(err)
		return
	}

	details := make([]detail, len(ratings))
	for i, rating := range ratings {
		details[i] = detail{ID: rating.Key.Parent.ID, GLO: rating.Rank().GLO()}
	}

	c.JSON(http.StatusOK, gin.H{"details": details})
}

func (g *Game) cache() (*datastore.Key, interface{}) {
	return newGameKey(g.id(), g.Undo.Current), g
}

func (g *Game) save() ([]*datastore.Key, []interface{}) {
	gh := newGHeader(g.id())
	gh.Header = g.Header

	ks := []*datastore.Key{newGCommittedKey(g.id()), newGameKey(g.id(), g.Undo.Current), gh.Key}
	es := []interface{}{g, g, gh}
	return ks, es
}

func (cl *client) drop(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	inv, err := cl.getInvitation()
	if err != nil {
		cl.jerr(err)
		return
	}

	cu := cl.CUser()
	if cu == nil {
		cl.jerr(user.ErrNotFound)
		return
	}

	err = inv.Drop(cu)
	if err != nil {
		cl.jerr(err)
		return
	}

	if len(inv.UserKeys) == 0 {
		inv.Status = game.Aborted
	}

	_, err = cl.DS.Put(c, inv.Key, inv)
	if err != nil {
		cl.jerr(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"invitation": inv,
		"message":    fmt.Sprintf("%s dropped from game invitation: %d", cu.Name, inv.Key.ID),
	})
}

func (cl *client) getInvitation() (*invitation, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	id, err := cl.getID()
	if err != nil {
		return nil, err
	}

	inv := newInvitation(id)
	err = cl.DS.Get(cl.ctx, inv.Key, inv)
	return inv, err
}

func (cl *client) getGCommited() error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	id, err := cl.getID()
	if err != nil {
		return err
	}

	cl.gc = newGCommited(id)
	return cl.DS.Get(cl.ctx, cl.gc.Key, cl.gc)
}

func (cl *client) getStack() (*undo.Stack, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	bodyBytes, err := copyBody(cl.ctx)
	if err != nil {
		return nil, err
	}

	obj := struct {
		*undo.Stack `json:"undo"`
	}{}
	err = cl.ctx.ShouldBind(&obj)
	if err != nil {
		return nil, err
	}

	// Restore the io.ReadCloser to its original state
	cl.ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
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

func (cl *client) invitationsIndex(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu := cl.CUser()
	if cu == nil {
		sn.JErr(c, user.ErrNotFound)
		return
	}

	q := datastore.
		NewQuery(invitationKind).
		Filter("Status=", int(game.Recruiting)).
		Order("-UpdatedAt")

	var es []*invitation
	_, err := cl.DS.GetAll(c, q, &es)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"invitations": es, "cu": cu})
}

func (cl *client) gamesIndex(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu := cl.CUser()
	if cu == nil {
		sn.JErr(c, user.ErrNotFound)
		return
	}

	status := game.ToStatus[c.Param("status")]
	q := datastore.
		NewQuery(headerKind).
		Filter("Status=", int(status)).
		Order("-UpdatedAt")

	var es []*GHeader
	_, err := cl.DS.GetAll(c, q, &es)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"gheaders": es, "cu": cu})
}

func (cl *client) cuHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.ctx = c

	cl.CUser()
	if cl.cu == nil {
		cl.jerr(user.ErrNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"cu": cl.cu})
}

// func jerr(c *gin.Context, err error) {
// 	if errors.Is(err, sn.ErrValidation) {
// 		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
// 		return
// 	}
// 	log.Debugf(err.Error())
// 	c.JSON(http.StatusOK, gin.H{"message": sn.ErrUnexpected.Error()})
// }

func (cl *client) jerr(err error) {
	if errors.Is(err, sn.ErrValidation) {
		cl.ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	cl.Log.Errorf(err.Error())
	cl.ctx.JSON(http.StatusOK, gin.H{"message": sn.ErrUnexpected.Error()})
}
