package got

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/codec"
	"github.com/SlothNinja/color"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/memcache"
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
		cu := user.CurrentFrom(c)
		c.HTML(http.StatusOK, prefix+"/show", gin.H{
			"Context":    c,
			"VersionID":  sn.VersionID(),
			"CUser":      cu,
			"Game":       g,
			"IsAdmin":    user.IsAdmin(c),
			"Admin":      game.AdminFrom(c),
			"MessageLog": mlog.From(c),
			"ColorMap":   color.MapFrom(c),
		})
	}
}

func (g *Game) update(c *gin.Context) (tmpl string, act game.ActionType, err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	a := c.PostForm("action")
	log.Debugf("action: %#v", a)
	switch a {
	case "select-area":
		tmpl, act, err = g.selectArea(c)
	case "admin-header":
		tmpl, act, err = g.adminHeader(c)
	case "admin-player":
		tmpl, act, err = g.adminPlayer(c)
	case "pass":
		tmpl, act, err = g.pass(c)
	case "undo":
		tmpl, act, err = g.undoTurn(c)
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
			log.Errorf("Controller#Update Game Not Found")
			return
		}

		mkey := g.UndoKey(c)
		if err := memcache.Delete(c, mkey); err != nil && err != memcache.ErrCacheMiss {
			log.Errorf("Controller#Undo Error: %s", err)
		}
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
		template, actionType, err := g.update(c)
		switch {
		case err != nil && sn.IsVError(err):
			restful.AddErrorf(c, "%v", err)
			withJSON(c, g)
		case err != nil:
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		case actionType == game.Cache:
			mkey := g.UndoKey(c)
			item := &memcache.Item{
				Key:        mkey,
				Expiration: time.Minute * 30,
			}

			v, err := codec.Encode(g)
			if err != nil {
				log.Errorf(err.Error())
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}
			item.Value = v
			if err := memcache.Set(c, item); err != nil {
				log.Errorf(err.Error())
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}
		case actionType == game.Save:
			if err := client.save(c, g); err != nil {
				log.Errorf("%s", err)
				restful.AddErrorf(c, "Controller#Update Save Error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}
		case actionType == game.Undo:
			mkey := g.UndoKey(c)
			if err := memcache.Delete(c, mkey); err != nil && err != memcache.ErrCacheMiss {
				log.Errorf("memcache.Delete error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
			}
		}

		switch jData := jsonFrom(c); {
		case jData != nil && template == "json":
			c.JSON(http.StatusOK, jData)
		case template == "":
			c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
		default:
			cu := user.CurrentFrom(c)
			d := gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     cu,
				"Game":      g,
				"IsAdmin":   user.IsAdmin(c),
				"Notices":   restful.NoticesFrom(c),
				"Errors":    restful.ErrorsFrom(c),
			}
			log.Debugf("d: %#v", d)
			c.HTML(http.StatusOK, template, d)
		}
	}
}

func (client Client) save(c *gin.Context, g *Game) error {
	oldG := New(c, g.ID())
	err := client.Get(c, oldG.Header.Key, oldG.Header)
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

	_, err = client.Put(c, g.Key, g.Header)
	if err != nil {
		return err
	}

	err = memcache.Delete(c, g.UndoKey(c))
	if err == memcache.ErrCacheMiss {
		return nil
	}
	return err
}

func (client Client) saveWith(c *gin.Context, g *Game, ks []*datastore.Key, es []interface{}) error {
	_, err := client.RunInTransaction(c, func(tx *datastore.Transaction) error {
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

		err = memcache.Delete(c, g.UndoKey(c))
		if err == memcache.ErrCacheMiss {
			return nil
		}
		return err
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

	g.TempData = nil

	var encoded []byte
	if encoded, err = codec.Encode(g.State); err != nil {
		return
	}
	g.SavedState = encoded
	g.updateHeader()

	return
}

//func (g *Game) saveAndUpdateStats(c *gin.Context) error {
//	ctx := restful.ContextFrom(c)
//	cu := user.CurrentFrom(c)
//	s, err := stats.ByUser(c, cu)
//	if err != nil {
//		return err
//	}
//
//	return datastore.RunInTransaction(ctx, func(tc context.Context) error {
//		c = restful.WithContext(c, tc)
//		oldG := New(c)
//		if ok := datastore.PopulateKey(oldG.Header, datastore.KeyForObj(tc, g.Header)); !ok {
//			return fmt.Errorf("Unable to populate game with key.")
//		}
//		if err := datastore.Get(tc, oldG.Header); err != nil {
//			return err
//		}
//
//		if oldG.UpdatedAt != g.UpdatedAt {
//			return fmt.Errorf("Game state changed unexpectantly.  Try again.")
//		}
//
//		g.TempData = nil
//		if encoded, err := codec.Encode(g.State); err != nil {
//			return err
//		} else {
//			g.SavedState = encoded
//		}
//
//		es := []interface{}{s, g.Header}
//		if err := datastore.Put(tc, es); err != nil {
//			return err
//		}
//		if err := memcache.Delete(tc, g.UndoKey(c)); err != nil && err != memcache.ErrCacheMiss {
//			return err
//		}
//		return nil
//	}, &datastore.TransactionOptions{XG: true})
//}

func newGamer(c *gin.Context) game.Gamer {
	return New(c, 0)
}

func (client Client) index(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		gs := game.GamersFrom(c)
		switch status := game.StatusFrom(c); status {
		case game.Recruiting:
			c.HTML(http.StatusOK, "shared/invitation_index", gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     user.CurrentFrom(c),
				"Games":     gs,
				"Type":      gtype.GOT.String(),
			})
		default:
			c.HTML(http.StatusOK, "shared/games_index", gin.H{
				"Context":   c,
				"VersionID": sn.VersionID(),
				"CUser":     user.CurrentFrom(c),
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
		err := g.FromParams(c, gtype.GOT)
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		c.HTML(http.StatusOK, prefix+"/new", gin.H{
			"Context":   c,
			"VersionID": sn.VersionID(),
			"CUser":     user.CurrentFrom(c),
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

		err := g.FromForm(c, g.Type)
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

		ks, err := client.AllocateIDs(c, []*datastore.Key{g.Header.Key})
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}
		g.Header.Key = ks[0]

		_, err = client.RunInTransaction(c, func(tx *datastore.Transaction) error {
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

		u := user.CurrentFrom(c)
		start, err := g.Accept(c, u)
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

		err = client.save(c, g)
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

		u := user.CurrentFrom(c)
		err := g.Drop(u)
		if err != nil {
			log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
		}
		err = client.save(c, g)

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

	g, err := client.getGame(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	withGame(c, g)
	cm := g.ColorMapFor(user.CurrentFrom(c))
	color.WithMap(c, cm)
}

func (client Client) getGame(c *gin.Context) (*Game, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	id, err := strconv.ParseInt(c.Param("hid"), 10, 64)
	if err != nil {
		return nil, err
	}

	g := New(c, id)
	switch c.PostForm("action") {
	case "reset", "undo":
		return client.dsGet(c, g)
	default:
		if user.CurrentFrom(c) != nil {
			g, err = client.mcGet(c, g)
			if err == nil {
				return g, nil
			}
		}

		return client.dsGet(c, g)
	}
}

// pull temporary game state from memcache.  Note may be different from value stored in datastore.
func (client Client) mcGet(c *gin.Context, g *Game) (*Game, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	mkey := g.GetHeader().UndoKey(c)
	item, err := memcache.Get(c, mkey)
	if err != nil {
		return nil, err
	}

	err = codec.Decode(g, item.Value)
	if err != nil {
		return nil, err
	}

	err = client.afterCache(c, g)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// pull game state from memcache/datastore.  returned memcache should be same as datastore.
func (client Client) dsGet(c *gin.Context, g *Game) (*Game, error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	err := client.Get(c, g.Header.Key, g.Header)
	switch {
	case err != nil:
		restful.AddErrorf(c, err.Error())
		return nil, err
	case g == nil:
		err = fmt.Errorf("Unable to get game for id: %v", g.ID)
		restful.AddErrorf(c, err.Error())
		return nil, err
	}

	state := newState()
	err = codec.Decode(&state, g.SavedState)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return nil, err
	}

	g.State = state

	err = client.init(c, g)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return nil, err
	}

	return g, nil
}

func (client Client) jsonIndexAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		client.Game.JSONIndexAction(c, gtype.GOT)
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
