package main

import (
	"bytes"
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

var (
	ErrNotFound     = fmt.Errorf("not found: %w", sn.ErrValidation)
	ErrInvalidCache = fmt.Errorf("invalid cached item: %w", sn.ErrValidation)
)

func (cl *client) homeHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu, err := cl.User.Current(c)
	if err != nil {
		cl.Log.Warningf(err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"cu": cu})
}

func (cl *client) showHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	gc, err := cl.getGCommited(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cu, err := cl.User.Current(c)
	if err != nil {
		cl.Log.Warningf(err.Error())
	}

	g := &(gc.Game)
	g.updateClickablesFor(cu, g.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{
		"game": g,
		"cu":   cu,
	})
}

func (cl *client) mlogHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	id, err := getID(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	ml, err := cl.MLog.Get(c, id)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": ml.Messages})
}

func (cl *client) mlogAddHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu, err := cl.User.Current(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	id, err := getID(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	obj := struct {
		Message string     `json:"message"`
		Creator *user.User `json:"creator"`
	}{}

	err = c.ShouldBind(&obj)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if obj.Creator.ID() != cu.ID() {
		sn.JErr(c, fmt.Errorf("invalid creator: %w", sn.ErrValidation))
		return
	}

	ml, err := cl.MLog.Get(c, id)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	m := ml.AddMessage(cu, obj.Message)

	_, err = cl.MLog.Put(c, id, ml)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": m})
}

func (cl *client) undo(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g, cu, err := cl.getGame(c, (*undo.Stack).Undo)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	g.updateClickablesFor(cu, g.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (cl *client) redo(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g, cu, err := cl.getGame(c, (*undo.Stack).Redo)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	g.updateClickablesFor(cu, g.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (cl *client) reset(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g, cu, err := cl.getGame(c, (*undo.Stack).Reset)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	g.updateClickablesFor(cu, g.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": g})
	// cu, err := cl.User.Current(c)
	// if err != nil {
	// 	sn.JErr(c, err)
	// 	return
	// }

	// gc, err := cl.getGCommited(c)
	// if err != nil {
	// 	sn.JErr(c, err)
	// 	return
	// }

	// g, _, err := cl.putCachedGame(c, &(gc.Game), gc.id(), gc.Undo.Current)
	// if err != nil {
	// 	sn.JErr(c, err)
	// 	return
	// }

	// g.updateClickablesFor(cu, g.selectedThiefArea())
	// c.JSON(http.StatusOK, gin.H{"game": g})
}

func (cl *client) undoOperations(c *gin.Context, action func(*undo.Stack) bool) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	undo, err := cl.getStack(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	id, err := getID(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	changed := action(undo)
	if !changed {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	g, err := cl.getCachedGame(c, id, undo.Current)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if undo.Committed != g.Undo.Committed {
		sn.JErr(c, fmt.Errorf("invalid game state"))
		return
	}

	cu, err := cl.User.Current(c)
	if err != nil {
		cl.Log.Warningf(err.Error())
	}

	g.updateClickablesFor(cu, g.selectedThiefArea())
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (cl *client) newInvitationHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu, err := cl.User.Current(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	inv := defaultInvitation()

	c.JSON(http.StatusOK, gin.H{"invitation": inv, "cu": cu})
}

func (cl *client) createHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu, err := cl.User.Current(c)
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
		m := mlog.New(inv.Key.ID)
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
		"cu":         cu,
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
		inv.PasswordHash = hashed
	}
	inv.AddCreator(cu)
	inv.TwoThiefVariant = obj.TwoThiefVariant
	inv.AddUser(cu)
	inv.Status = game.Recruiting
	inv.Type = gtype.GOT
	return nil
}

func (cl *client) acceptHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	inv, err := cl.getInvitation(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cu, err := cl.User.Current(c)
	if err != nil {
		sn.JErr(c, err)
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

	start, err := inv.AcceptWith(cu, []byte(obj.Password))
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

	cl.sendTurnNotificationsTo(g, g.currentPlayer())

	inv.Header = g.Header
	c.JSON(http.StatusOK, gin.H{
		"invitation": inv,
		"message": fmt.Sprintf(
			`<div>Game: %d has started.</div>
			<div></div>
			<div><strong>%s</strong> is start player.</div>`,
			inv.ID(), g.nameFor(g.currentPlayer())),
	})
}

type detail struct {
	ID  int64 `json:"id"`
	GLO int   `json:"glo"`
}

func (cl *client) details(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	inv, err := cl.getInvitation(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cu, err := cl.User.Current(c)
	if err != nil {
		sn.JErr(c, err)
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
		sn.JErr(c, err)
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

func (cl *client) commit(c *gin.Context, g *Game) error {
	g.Undo.Commit()
	ks, es := g.save()
	_, err := cl.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		_, err := tx.PutMulti(ks, es)
		return err
	})

	if err != nil {
		return err
	}
	cl.Cache.Delete(ks[1].Encode())
	return nil
}

func (cl *client) dropHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	inv, err := cl.getInvitation(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	cu, err := cl.User.Current(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	err = inv.Drop(cu)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	if len(inv.UserKeys) == 0 {
		inv.Status = game.Aborted
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

func (cl *client) getInvitation(c *gin.Context) (*invitation, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	id, err := getID(c)
	if err != nil {
		return nil, err
	}

	inv := newInvitation(id)
	err = cl.DS.Get(c, inv.Key, inv)
	return inv, err
}

func (cl *client) getGCommited(c *gin.Context) (*gcommitted, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	id, err := getID(c)
	if err != nil {
		return nil, err
	}

	gc := newGCommited(id)
	err = cl.DS.Get(c, gc.Key, gc)
	if err != nil {
		return nil, err
	}
	return gc, nil
}

func (cl *client) getStack(c *gin.Context) (*undo.Stack, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	bodyBytes, err := copyBody(c)
	if err != nil {
		return nil, err
	}

	obj := struct {
		*undo.Stack `json:"undo"`
	}{}
	err = c.ShouldBind(&obj)
	if err != nil {
		return nil, err
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

func (cl *client) invitationsIndexHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu, err := cl.User.Current(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	q := datastore.
		NewQuery(invitationKind).
		Filter("Status=", int(game.Recruiting)).
		Order("-UpdatedAt")

	var es []*invitation
	_, err = cl.DS.GetAll(c, q, &es)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"invitations": es, "cu": cu})
}

func (cl *client) gamesIndex(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu, err := cl.User.Current(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	status := game.ToStatus[c.Param("status")]
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

func (cl *client) cuHandler(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cu, err := cl.User.Current(c)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"cu": cu})
}

func (cl *client) getGame(c *gin.Context, action ...func(*undo.Stack) bool) (*Game, *user.User, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	id, err := getID(c)
	if err != nil {
		return nil, nil, err
	}

	cu, err := cl.User.Current(c)
	if err != nil {
		cl.Log.Warningf(err.Error())
	}

	undo, err := cl.getStack(c)
	if err != nil {
		return nil, nil, err
	}

	// if no undo operation, pull current state of game
	if len(action) != 1 {
		g, err := cl.getCachedGame(c, id, undo.Current)
		if err != nil {
			return nil, nil, err
		}
		return g, cu, nil
	}

	// if an undo operation, verify user logged in
	if cu == nil {
		return nil, nil, fmt.Errorf("must be logged in: %w", sn.ErrValidation)
	}

	// if undo operation does not transistion to different state, pull current state of game
	if changed := action[0](undo); !changed {
		g, err := cl.getCachedGame(c, id, undo.Current)
		if err != nil {
			return nil, nil, err
		}
		return g, cu, nil
	}

	// Otherwise need to verify current user is current player or admin, which requires
	// getting the commited game state
	gc, err := cl.getGCommited(c)
	if err != nil {
		return nil, nil, err
	}

	_, err = gc.validateCPorAdmin(cu)
	if err != nil {
		return nil, nil, err
	}

	g, err := cl.getCachedGame(c, id, undo.Current)
	if err != nil {
		return nil, nil, err
	}
	g.Undo = *undo
	return g, cu, nil
}

func (cl *client) getCachedGame(c *gin.Context, id, rev int64) (*Game, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g, err := cl.mcGame(id, rev)
	if err == nil {
		return g, nil
	}

	return cl.dsGame(c, id, rev)
}

func (cl *client) mcGame(id, rev int64) (*Game, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g := newGame(id, rev)
	k := g.Key.Encode()
	item, found := cl.Cache.Get(k)
	if !found {
		return nil, ErrNotFound
	}

	ps, ok := item.([]datastore.Property)
	if !ok {
		cl.Cache.Delete(k)
		return nil, ErrInvalidCache
	}

	err := g.Load(ps)
	if err != nil {
		cl.Cache.Delete(k)
		return nil, err
	}

	return g, nil
}

func (cl *client) dsGame(c *gin.Context, id, rev int64) (*Game, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	g := newGame(id, rev)
	err := cl.DS.Get(c, g.Key, g)
	if err != nil {
		return nil, err
	}
	ps, err := g.Save()
	if err != nil {
		cl.Log.Warningf(err.Error())
		return g, nil
	}
	cl.Cache.SetDefault(g.Key.Encode(), ps)
	return g, nil
}

func (cl *client) putCachedGame(c *gin.Context, g *Game, id, rev int64) (*Game, *datastore.Key, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	k, err := cl.DS.Put(c, newGameKey(id, rev), g)
	if err != nil {
		return nil, nil, err
	}
	cl.Cache.Delete(k.Encode())
	return g, k, nil
}
