package got

import (
	"fmt"
	"net/http"
	"strconv"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/codec"
	"github.com/SlothNinja/color"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

const (
	gameKey   = "Game"
	homePath  = "/"
	jsonKey   = "JSON"
	statusKey = "Status"
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

func (client Client) show(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		cu, err := client.User.Current(c)
		if err != nil {
			log.Debugf(err.Error())
		}

		c.HTML(http.StatusOK, prefix+"/show", gin.H{
			"Context":    c,
			"VersionID":  sn.VersionID(),
			"CUser":      cu,
			"Game":       g,
			"IsAdmin":    cu.IsAdmin(),
			"Admin":      game.AdminFrom(c),
			"MessageLog": mlog.From(c),
			"ColorMap":   color.MapFrom(c),
		})
	}
}

func (g *Game) update(c *gin.Context, cu *user.User) (tmpl string, act game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	a := c.PostForm("action")
	log.Debugf("action: %#v", a)
	switch a {
	case "select-area":
		tmpl, act, err = g.selectArea(c, cu)
	case "admin-header":
		tmpl, act, err = g.adminHeader(c, cu)
	case "admin-player":
		tmpl, act, err = g.adminPlayer(c, cu)
	case "pass":
		tmpl, act, err = g.pass(c, cu)
	case "undo":
		tmpl, act, err = g.undoTurn(c, cu)
	default:
		act, err = game.None, fmt.Errorf("%v is not a valid action", a)
	}
	return
}

func (client Client) undo(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		if g == nil {
			log.Errorf("game not found")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "game not found"})
			return
		}

		cu, err := client.User.Current(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
		mkey := g.UndoKey(c, cu)
		client.Cache.Delete(mkey)
		c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
	}
}

func (client Client) update(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		if g == nil {
			log.Errorf("Controller#Update Game Not Found")
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}

		cu, err := client.User.Current(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
		template, actionType, err := g.update(c, cu)
		switch {
		case err != nil && sn.IsVError(err):
			restful.AddErrorf(c, "%v", err)
			withJSON(c, g)
		case err != nil:
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		case actionType == game.Cache:
			mkey := g.UndoKey(c, cu)
			client.Cache.SetDefault(mkey, g)
		case actionType == game.SaveAndStatUpdate:
			st, err := client.Stats.ByUser(c, cu)
			if err != nil {
				log.Errorf("stat.ByUser error: %v", err)
				restful.AddErrorf(c, "stats.ByUser error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}

			ks := []*datastore.Key{st.Key}
			es := []interface{}{st}

			err = client.saveWith(c, g, cu, ks, es)
			if err != nil {
				log.Errorf("g.save error: %s", err)
				restful.AddErrorf(c, "g.save error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}
		case actionType == game.Save:
			if err := client.save(c, g, cu); err != nil {
				log.Errorf("%s", err)
				restful.AddErrorf(c, "Controller#Update Save Error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}
		case actionType == game.Undo:
			mkey := g.UndoKey(c, cu)
			client.Cache.Delete(mkey)
		}

		switch jData := jsonFrom(c); {
		case jData != nil && template == "json":
			c.JSON(http.StatusOK, jData)
		case template == "":
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
		default:
			d := gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     cu,
				"Game":      g,
				"IsAdmin":   cu.IsAdmin(),
				"Notices":   restful.NoticesFrom(c),
				"Errors":    restful.ErrorsFrom(c),
			}
			log.Debugf("d: %#v", d)
			c.HTML(http.StatusOK, template, d)
		}
	}
}

func (client Client) save(c *gin.Context, g *Game, cu *user.User) error {
	oldG := New(c, g.ID())
	err := client.DS.Get(c, oldG.Header.Key, oldG.Header)
	if err != nil {
		return err
	}

	if oldG.UpdatedAt != g.UpdatedAt {
		return fmt.Errorf("game state changed unexpectantly -- try again")
	}

	err = g.encode(c)
	if err != nil {
		return err
	}

	_, err = client.DS.Put(c, g.Key, g.Header)
	if err != nil {
		return err
	}

	client.Cache.Delete(g.UndoKey(c, cu))
	return nil
}

func (client Client) saveWith(c *gin.Context, g *Game, cu *user.User, ks []*datastore.Key, es []interface{}) error {
	_, err := client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		oldG := New(c, g.ID())
		err := tx.Get(oldG.Header.Key, oldG.Header)
		if err != nil {
			return err
		}

		if oldG.UpdatedAt != g.UpdatedAt {
			return fmt.Errorf("game state changed unexpectantly -- try again")
		}

		err = g.encode(c)
		if err != nil {
			return err
		}

		ks = append(ks, g.Key)
		es = append(es, g.Header)

		_, err = tx.PutMulti(ks, es)
		if err != nil {
			return err
		}

		client.Cache.Delete(g.UndoKey(c, cu))
		return nil
	})
	return err
}

func wrap(s *stats.Stats, cs contest.Contests) ([]*datastore.Key, []interface{}) {
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

func (g *Game) encode(c *gin.Context) (err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	var encoded []byte
	if encoded, err = codec.Encode(g.State); err != nil {
		return
	}
	g.SavedState = encoded
	g.updateHeader()

	return
}

func newGamer(c *gin.Context) game.Gamer {
	return New(c, 0)
}

func (client Client) index(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		gs := game.GamersFrom(c)
		cu, err := client.User.Current(c)
		if err != nil {
			log.Debugf(err.Error())
		}
		switch status := game.StatusFrom(c); status {
		case game.Recruiting:
			c.HTML(http.StatusOK, "shared/invitation_index", gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     cu,
				"Games":     gs,
				"Type":      gtype.GOT.String(),
			})
		default:
			c.HTML(http.StatusOK, "shared/games_index", gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     cu,
				"Games":     gs,
				"Type":      gtype.GOT.String(),
				"Status":    status,
			})
		}
	}
}

//func Index(prefix string) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		ctx := restful.ContextFrom(c)
//		log.Debugf(ctx, "Entering")
//		defer log.Debugf(ctx, "Exiting")
//
//		gs := game.GamersFrom(ctx)
//		switch {
//		case game.StatusFrom(ctx) == game.Recruiting:
//			c.HTML(http.StatusOK, "shared/invitation_index", gin.H{
//				"Context":   ctx,
//				"VersionID": appengine.VersionID(ctx),
//				"CUser":     user.CurrentFrom(ctx),
//				"Games":     gs,
//			})
//		case gtype.TypeFrom(ctx) == gtype.All:
//			c.HTML(http.StatusOK, "shared/multi_games_index", gin.H{
//				"Context":   ctx,
//				"VersionID": appengine.VersionID(ctx),
//				"CUser":     user.CurrentFrom(ctx),
//				"Games":     gs,
//			})
//		default:
//			c.HTML(http.StatusOK, "shared/games_index", gin.H{
//				"Context":   ctx,
//				"VersionID": appengine.VersionID(ctx),
//				"CUser":     user.CurrentFrom(ctx),
//				"Games":     gs,
//			})
//		}
//	}
//}

func recruitingPath(prefix string) string {
	return fmt.Sprintf("/%s/games/recruiting", prefix)
}

func newPath(prefix string) string {
	return fmt.Sprintf("/%s/game/new", prefix)
}

func (client Client) newAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := New(c, 0)
		withGame(c, g)
		cu, err := client.User.Current(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.FromParams(c, cu, gtype.GOT)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		c.HTML(http.StatusOK, prefix+"/new", gin.H{
			"Context":   c,
			"VersionID": sn.VersionID(),
			"CUser":     cu,
			"Game":      g,
		})
	}
}

func (client Client) create(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := New(c, 0)
		withGame(c, g)

		cu, err := client.User.Current(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.FromForm(c, cu, g.Type)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.fromForm(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.encode(c)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		log.Debugf("before client.DS.AllocateIDs")
		ks, err := client.DS.AllocateIDs(c, []*datastore.Key{g.Header.Key})
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}
		g.Header.Key = ks[0]

		log.Debugf("before client.DS.RunInTransaction")
		_, err = client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
			m := mlog.New(g.Header.Key.ID)
			ks := []*datastore.Key{g.Header.Key, m.Key}
			es := []interface{}{g.Header, m}

			_, err := tx.PutMulti(ks, es)
			return err
		})
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		restful.AddNoticef(c, "<div>%s created.</div>", g.Title)
		c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
		return
	}
}

func (client Client) accept(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		if g == nil {
			log.Errorf("game not found")
			defer c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		cu, err := client.User.Current(c)
		if err != nil {
			log.Debugf(err.Error())
		}
		start, err := g.Accept(c, cu)
		if err != nil {
			log.Errorf(err.Error())
			defer c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		if start {
			err = g.Start(c)
			if err != nil {
				log.Errorf(err.Error())
				defer c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
				return
			}
		}

		err = client.save(c, g, cu)
		if err != nil {
			log.Errorf(err.Error())
			defer c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		if start {
			err = g.SendTurnNotificationsTo(c, g.CurrentPlayer())
			if err != nil {
				log.Errorf(err.Error())
			}
		}
		defer c.Redirect(http.StatusSeeOther, recruitingPath(prefix))

	}
}

func (client Client) drop(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		g := gameFrom(c)
		if g == nil {
			log.Errorf("game not found")
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		cu, err := client.User.Current(c)
		if err != nil {
			log.Debugf(err.Error())
		}
		err = g.Drop(cu)
		if err != nil {
			log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
		}
		err = client.save(c, g, cu)

		if err != nil {
			log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
		}
		c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
	}
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
	cu, err := client.User.Current(c)
	if err != nil {
		log.Debugf(err.Error())
	}

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
		if cu != nil {
			// pull from cache and return if successful; otherwise pull from datastore
			err = client.mcGet(c, g, cu)
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
func (client Client) mcGet(c *gin.Context, g *Game, cu *user.User) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	mkey := g.UndoKey(c, cu)
	item, found := client.Cache.Get(mkey)
	if !found {
		return fmt.Errorf("not found")
	}

	g2, ok := item.(*Game)
	if !ok {
		return fmt.Errorf("item in cache is not a *Game")
	}
	g2.SetCTX(c)

	g = g2

	withGame(c, g)
	cu, err := client.User.Current(c)
	if err != nil {
		log.Debugf(err.Error())
	}
	cm := g.ColorMapFor(cu)
	color.WithMap(c, cm)
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

	state := newState()
	err = codec.Decode(&state, g.SavedState)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return err
	}

	g.State = state

	err = client.init(c, g)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return err
	}

	withGame(c, g)
	cu, err := client.User.Current(c)
	if err != nil {
		log.Debugf(err.Error())
	}
	cm := g.ColorMapFor(cu)
	color.WithMap(c, cm)
	return nil
}

func (client Client) jsonIndexAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		client.Game.JSONIndexAction(c)
	}
}

func (g *Game) updateHeader() {
	g.OptString = g.options()
	switch g.Phase {
	case gameOver:
		g.Progress = g.PhaseName()
	default:
		g.Progress = fmt.Sprintf("<div>Turn: %d</div><div>Phase: %s</div>", g.Turn, g.PhaseName())
	}
	if u := g.Creator; u != nil {
		g.CreatorSID = user.GenID(u.GoogleID)
		g.CreatorName = u.Name
	}

	if l := len(g.Users); l > 0 {
		g.UserSIDS = make([]string, l)
		g.UserNames = make([]string, l)
		g.UserEmails = make([]string, l)
		for i, u := range g.Users {
			g.UserSIDS[i] = user.GenID(u.GoogleID)
			g.UserNames[i] = u.Name
			g.UserEmails[i] = u.Email
		}
	}
}
