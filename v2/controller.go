package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"cloud.google.com/go/datastore"
	"github.com/Pallinder/go-randomdata"
	"github.com/SlothNinja/codec"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
)

const (
	gameKey        = "Game"
	jsonKey        = "JSON"
	statusKey      = "Status"
	homePath       = "/"
	recruitingPath = "/games/recruiting"
	newPath        = "/game/new"
)

func gameFrom(c *gin.Context) (g *Game) {
	g, _ = c.Value(gameKey).(*Game)
	return
}

func withGame(c *gin.Context, g *Game) {
	c.Set(gameKey, g)
}

func jsonFrom(c *gin.Context) (g *Game) {
	g, _ = c.Value(jsonKey).(*Game)
	return
}

func withJSON(c *gin.Context, g *Game) {
	c.Set(jsonKey, g)
}

func (client Client) show(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	// g := gameFrom(c)
	// cu := user.CurrentFrom(c)
	// c.HTML(http.StatusOK, "/show", gin.H{
	// 	"Context":    c,
	// 	"VersionID":  sn.VersionID(),
	// 	"CUser":      cu,
	// 	"Game":       g,
	// 	"IsAdmin":    user.IsAdmin(c),
	// 	"Admin":      game.AdminFrom(c),
	// 	"MessageLog": mlog.From(c),
	// 	"ColorMap":   color.MapFrom(c),
	// })
}

func (g *Game) update(c *gin.Context) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	// a := c.PostForm("action")
	// log.Debugf("action: %#v", a)
	// switch a {
	// case "select-area":
	// 	tmpl, act, err = g.selectArea(c)
	// case "admin-header":
	// 	tmpl, act, err = g.adminHeader(c)
	// case "admin-player":
	// 	tmpl, act, err = g.adminPlayer(c)
	// case "pass":
	// 	tmpl, act, err = g.pass(c)
	// case "undo":
	// 	tmpl, act, err = g.undoTurn(c)
	// default:
	// 	act, err = game.None, fmt.Errorf("%v is not a valid action", a)
	// }
	return nil
}

func (client Client) undo(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	// g := gameFrom(c)
	// if g == nil {
	// 	log.Errorf("game not found")
	// 	c.JSON(http.StatusInternalServerError, gin.H{"Error": "game not found"})
	// 	return
	// }

	// mkey := g.UndoKey(c)
	// client.Cache.Delete(mkey)
	// c.Redirect(http.StatusSeeOther, showPath(c.Param("hid")))
}

func (client Client) update(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	// g := gameFrom(c)
	// if g == nil {
	// 	log.Errorf("Controller#Update Game Not Found")
	// 	c.Redirect(http.StatusSeeOther, homePath)
	// 	return
	// }
	// template, actionType, err := g.update(c)
	// switch {
	// case err != nil && sn.IsVError(err):
	// 	restful.AddErrorf(c, "%v", err)
	// 	withJSON(c, g)
	// case err != nil:
	// 	log.Errorf(err.Error())
	// 	c.Redirect(http.StatusSeeOther, homePath)
	// 	return
	// case actionType == game.Cache:
	// 	mkey := g.UndoKey(c)
	// 	client.Cache.SetDefault(mkey, g)
	// case actionType == game.SaveAndStatUpdate:
	// 	cu := user.CurrentFrom(c)
	// 	st, err := client.Stats.ByUser(c, cu)
	// 	if err != nil {
	// 		log.Errorf("stat.ByUser error: %v", err)
	// 		restful.AddErrorf(c, "stats.ByUser error: %s", err)
	// 		c.Redirect(http.StatusSeeOther, showPath(c.Param("hid")))
	// 		return
	// 	}

	// 	ks := []*datastore.Key{st.Key}
	// 	es := []interface{}{st}

	// 	err = client.saveWith(c, g, ks, es)
	// 	if err != nil {
	// 		log.Errorf("g.save error: %s", err)
	// 		restful.AddErrorf(c, "g.save error: %s", err)
	// 		c.Redirect(http.StatusSeeOther, showPath(c.Param("hid")))
	// 		return
	// 	}
	// case actionType == game.Save:
	// 	if err := client.save(c, g); err != nil {
	// 		log.Errorf("%s", err)
	// 		restful.AddErrorf(c, "Controller#Update Save Error: %s", err)
	// 		c.Redirect(http.StatusSeeOther, showPath(c.Param("hid")))
	// 		return
	// 	}
	// case actionType == game.Undo:
	// 	mkey := g.UndoKey(c)
	// 	client.Cache.Delete(mkey)
	// }

	// switch jData := jsonFrom(c); {
	// case jData != nil && template == "json":
	// 	c.JSON(http.StatusOK, jData)
	// case template == "":
	// 	c.Redirect(http.StatusSeeOther, showPath(c.Param("hid")))
	// default:
	// 	cu := user.CurrentFrom(c)
	// 	d := gin.H{
	// 		"Context":   c,
	// 		"VersionID": sn.VersionID(),
	// 		"CUser":     cu,
	// 		"Game":      g,
	// 		"IsAdmin":   user.IsAdmin(c),
	// 		"Notices":   restful.NoticesFrom(c),
	// 		"Errors":    restful.ErrorsFrom(c),
	// 	}
	// 	log.Debugf("d: %#v", d)
	// 	c.HTML(http.StatusOK, template, d)
	// }
}

func (client Client) save(c *gin.Context, g *Game) error {
	// oldG := New(c, g.ID())
	// err := client.DS.Get(c, oldG.Header.Key, oldG.Header)
	// if err != nil {
	// 	return err
	// }

	// if oldG.UpdatedAt != g.UpdatedAt {
	// 	return fmt.Errorf("game state changed unexpectantly -- try again")
	// }

	// err = g.encode(c)
	// if err != nil {
	// 	return err
	// }

	// _, err = client.DS.Put(c, g.Key, g.Header)
	// if err != nil {
	// 	return err
	// }

	// client.Cache.Delete(g.UndoKey(c))
	return nil
}

func (client Client) saveWith(c *gin.Context, g *Game, ks []*datastore.Key, es []interface{}) error {
	// _, err := client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
	// 	oldG := New(c, g.ID())
	// 	err := tx.Get(oldG.Header.Key, oldG.Header)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if oldG.UpdatedAt != g.UpdatedAt {
	// 		return fmt.Errorf("game state changed unexpectantly -- try again")
	// 	}

	// 	err = g.encode(c)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	ks = append(ks, g.Key)
	// 	es = append(es, g.Header)

	// 	_, err = tx.PutMulti(ks, es)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	client.Cache.Delete(g.UndoKey(c))
	// 	return nil
	// })
	return nil
}

func wrap(s *user.Stats, cs []*sn.Contest) ([]*datastore.Key, []interface{}) {
	l := len(cs) + 1
	es := make([]interface{}, l)
	ks := make([]*datastore.Key, l)
	es[0] = s
	ks[0] = s.Key
	for i, c := range cs {
		es[i+1] = c
		ks[i+1] = c.Key
	}
	return ks, es
}

func (g *Game) encode(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	encoded, err := codec.Encode(g.State)
	if err != nil {
		return nil
	}
	g.SavedState = encoded

	return nil
}

// func newGamer(c *gin.Context) game.Gamer {
// 	return New(c, 0)
// }

func (client Client) index(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	// gs := game.GamersFrom(c)
	// switch status := game.StatusFrom(c); status {
	// case game.Recruiting:
	// 	c.HTML(http.StatusOK, "shared/invitation_index", gin.H{
	// 		"Context":   c,
	// 		"VersionID": sn.VersionID(),
	// 		"CUser":     user.CurrentFrom(c),
	// 		"Games":     gs,
	// 		"Type":      gtype.GOT.String(),
	// 	})
	// default:
	// 	c.HTML(http.StatusOK, "shared/games_index", gin.H{
	// 		"Context":   c,
	// 		"VersionID": sn.VersionID(),
	// 		"CUser":     user.CurrentFrom(c),
	// 		"Games":     gs,
	// 		"Type":      gtype.GOT.String(),
	// 		"Status":    status,
	// 	})
	// }
}

func (client Client) newAction(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cu, err := user.FromSession(c)
	if err != nil || cu == nil {
		jerr(c, sn.ErrUserNotFound)
		return
	}

	g := New(c, 0)

	// Default Values
	g.Title = fmt.Sprintf("%s's %s", cu.Name, randomdata.SillyName())
	g.NumPlayers = 2
	g.TwoThiefVariant = false

	c.JSON(http.StatusOK, gin.H{"header": g.Header})
}

// func (s server) newInvitation() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		log.Debugf(msgEnter)
// 		defer log.Debugf(msgExit)
//
// 		cu := user.Current(c)
// 		if cu == user.None {
// 			jerr(c, errUserNotFound)
// 			return
// 		}
//
// 		e := newHeaderEntity(newGame(0))
//
// 		// Default Values
// 		e.Title = fmt.Sprintf("%s's %s", cu.Name, randomdata.SillyName())
// 		e.NumPlayers = 2
// 		e.TwoThiefVariant = false
//
// 		c.JSON(http.StatusOK, gin.H{"header": e})
// 	}
// }

func (client Client) create(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	cu, err := user.FromSession(c)
	if err != nil {
		jerr(c, err)
		return
	}

	g := New(c, 0)
	err = g.fromForm(c, cu)
	if err != nil {
		jerr(c, err)
		return
	}

	err = g.encode(c)
	if err != nil {
		jerr(c, err)
		return
	}

	ks, err := client.DS.AllocateIDs(c, []*datastore.Key{g.Header.Key})
	if err != nil {
		jerr(c, err)
		return
	}
	g.Header.Key = ks[0]

	_, err = client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		m := sn.NewMLog(g.Header.Key.ID)
		ks := []*datastore.Key{g.Header.Key, m.Key}
		es := []interface{}{g.Header, m}

		_, err := tx.PutMulti(ks, es)
		return err
	})
	if err != nil {
		jerr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s created game %q", cu.Name, g.Title)})
	return
}

func (g *Game) fromForm(c *gin.Context, cu *user.User) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	obj := struct {
		Title           string `form:"title"`
		NumPlayers      int    `form:"num-players" binding"min=0,max=5"`
		Password        string `form:"password"`
		TwoThiefVariant bool   `form:"two-thief-variant"`
	}{}

	err := c.ShouldBind(&obj)
	if err != nil {
		return err
	}

	g.Title = cu.Name + "'s Game"
	if obj.Title != "" {
		g.Title = obj.Title
	}

	g.NumPlayers = 4
	if obj.NumPlayers >= 1 && obj.NumPlayers <= 5 {
		g.NumPlayers = obj.NumPlayers
	}

	g.Password = obj.Password
	g.AddCreator(cu)
	g.TwoThiefVariant = obj.TwoThiefVariant
	g.AddUser(cu)
	g.Status = sn.Recruiting
	return nil
}

func (client Client) accept(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g := gameFrom(c)
	if g == nil {
		log.Errorf("game not found")
		defer c.Redirect(http.StatusSeeOther, recruitingPath)
		return
	}

	u := user.CurrentFrom(c)
	start, err := g.Accept(c, u)
	if err != nil {
		log.Errorf(err.Error())
		defer c.Redirect(http.StatusSeeOther, recruitingPath)
		return
	}

	if start {
		err = g.Start(c)
		if err != nil {
			log.Errorf(err.Error())
			defer c.Redirect(http.StatusSeeOther, recruitingPath)
			return
		}
	}

	err = client.save(c, g)
	if err != nil {
		log.Errorf(err.Error())
		defer c.Redirect(http.StatusSeeOther, recruitingPath)
		return
	}

	if start {
		err = g.SendTurnNotificationsTo(c, g.CurrentPlayer())
		if err != nil {
			log.Errorf(err.Error())
		}
	}
	defer c.Redirect(http.StatusSeeOther, recruitingPath)

}

func (client Client) drop(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g := gameFrom(c)
	if g == nil {
		log.Errorf("game not found")
		c.Redirect(http.StatusSeeOther, recruitingPath)
		return
	}

	u := user.CurrentFrom(c)
	err := g.Drop(u)
	if err != nil {
		log.Errorf(err.Error())
		restful.AddErrorf(c, err.Error())
		c.Redirect(http.StatusSeeOther, recruitingPath)
	}
	err = client.save(c, g)

	if err != nil {
		log.Errorf(err.Error())
		restful.AddErrorf(c, err.Error())
		c.Redirect(http.StatusSeeOther, recruitingPath)
	}
	c.Redirect(http.StatusSeeOther, recruitingPath)
}

func (client Client) fetch(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")
	// create Gamer
	log.Debugf("hid: %v", c.Param("hid"))
	id, err := strconv.ParseInt(c.Param("hid"), 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Debugf("id: %v", id)
	g := New(c, id)

	switch action := c.PostForm("action"); {
	case action == "reset":
		// pull from cache/datastore
		// same as undo
		fallthrough
	case action == "undo":
		// pull from cache/datastore
		err = client.dsGet(c, g)
		if err != nil {
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	default:
		if user.CurrentFrom(c) != nil {
			// pull from cache and return if successful; otherwise pull from datastore
			err = client.mcGet(c, g)
			if err == nil {
				return
			}
		}

		err = client.dsGet(c, g)
		if err != nil {
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	}
}

// pull temporary game state from cache.  Note may be different from value stored in datastore.
func (client Client) mcGet(c *gin.Context, g *Game) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	mkey := g.UndoKey(c)
	item, found := client.Cache.Get(mkey)
	if !found {
		return fmt.Errorf("not found")
	}

	g2, ok := item.(*Game)
	if !ok {
		return fmt.Errorf("item in cache is not a *Game")
	}

	g = g2
	return nil
}

// pull game state from cache/datastore.  returned cache should be same as datastore.
func (client Client) dsGet(c *gin.Context, g *Game) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := client.DS.Get(c, g.Header.Key, g.Header)
	switch {
	case err != nil:
		restful.AddErrorf(c, err.Error())
		return err
	case g == nil:
		err = fmt.Errorf("Unable to get game for id: %v", g.ID)
		restful.AddErrorf(c, err.Error())
		return err
	}

	state := new(State)
	err = codec.Decode(&state, g.SavedState)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return err
	}

	g.State = state
	return nil
}

func (client Client) jsonIndexAction(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	cu, err := user.FromSession(c)
	if err != nil {
		jerr(c, err)
	}

	if cu == nil {
		jerr(c, sn.ErrUserNotFound)
		return
	}

	status := sn.ToStatus[c.Param("status")]
	q := datastore.
		NewQuery("Header").
		Filter("Status=", int(status)).
		Order("-UpdatedAt")

	var es []*Header
	_, err = client.DS.GetAll(c, q, &es)
	if err != nil {
		jerr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"headers": es, "cu": cu})
}

func (client Client) Current(c *gin.Context) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	u, err := user.FromSession(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cu": u})
}

// func (g *Game) updateHeader() {
// 	g.OptString = g.options()
// 	switch g.Phase {
// 	case gameOver:
// 		g.Progress = g.PhaseName()
// 	default:
// 		g.Progress = fmt.Sprintf("<div>Turn: %d</div><div>Phase: %s</div>", g.Turn, g.PhaseName())
// 	}
// 	if u := g.Creator; u != nil {
// 		g.CreatorSID = user.GenID(u.GoogleID)
// 		g.CreatorName = u.Name
// 	}
//
// 	if l := len(g.Users); l > 0 {
// 		g.UserSIDS = make([]string, l)
// 		g.UserNames = make([]string, l)
// 		g.UserEmails = make([]string, l)
// 		for i, u := range g.Users {
// 			g.UserSIDS[i] = user.GenID(u.GoogleID)
// 			g.UserNames[i] = u.Name
// 			g.UserEmails[i] = u.Email
// 		}
// 	}
// }

func jerr(c *gin.Context, err error) {
	if errors.Is(err, sn.ErrValidation) {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	log.Debugf(err.Error())
	c.JSON(http.StatusOK, gin.H{"message": sn.ErrUnexpected.Error()})
}
