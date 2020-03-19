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
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
	"google.golang.org/appengine"
	"google.golang.org/appengine/memcache"
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

func show(prefix string) gin.HandlerFunc {
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

func undo(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")
		c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))

		g := gameFrom(c)
		if g == nil {
			log.Errorf("Controller#Update Game Not Found")
			return
		}
		mkey := g.UndoKey(c)
		if err := memcache.Delete(appengine.NewContext(c.Request), mkey); err != nil && err != memcache.ErrCacheMiss {
			log.Errorf("Controller#Undo Error: %s", err)
		}
	}
}

func update(prefix string) gin.HandlerFunc {
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
			// item := memcache.NewItem(c, mkey).SetExpiration(time.Minute * 30)
			v, err := codec.Encode(g)
			if err != nil {
				log.Errorf(err.Error())
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}
			item.Value = v
			if err := memcache.Set(appengine.NewContext(c.Request), item); err != nil {
				log.Errorf(err.Error())
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}
		case actionType == game.SaveAndStatUpdate:
			cu := user.CurrentFrom(c)
			s, err := stats.ByUser(c, cu)
			if err != nil {
				log.Errorf("stat.ByUser error: %v", err)
				restful.AddErrorf(c, "stats.ByUser error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}

			ks := []*datastore.Key{s.Key}
			es := []interface{}{s}

			err = g.saveWith(c, ks, es)
			if err != nil {
				log.Errorf("g.save error: %s", err)
				restful.AddErrorf(c, "g.save error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}
		case actionType == game.Save:
			if err := g.save(c); err != nil {
				log.Errorf("%s", err)
				restful.AddErrorf(c, "Controller#Update Save Error: %s", err)
				c.Redirect(http.StatusSeeOther, showPath(prefix, c.Param("hid")))
				return
			}
		case actionType == game.Undo:
			mkey := g.UndoKey(c)
			if err := memcache.Delete(appengine.NewContext(c.Request), mkey); err != nil && err != memcache.ErrCacheMiss {
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

func (g *Game) save(c *gin.Context) error {
	dsClient, err := datastore.NewClient(c, "")
	if err != nil {
		return err
	}

	oldG := New(c, g.ID())
	err = dsClient.Get(c, oldG.Header.Key, oldG.Header)
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

	_, err = dsClient.Put(c, g.Key, g.Header)
	if err != nil {
		return err
	}

	err = memcache.Delete(appengine.NewContext(c.Request), g.UndoKey(c))
	if err == memcache.ErrCacheMiss {
		return nil
	}
	return err
}

func (g *Game) saveWith(c *gin.Context, ks []*datastore.Key, es []interface{}) error {
	dsClient, err := datastore.NewClient(c, "")
	if err != nil {
		return err
	}

	_, err = dsClient.RunInTransaction(c, func(tx *datastore.Transaction) error {
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

		err = memcache.Delete(appengine.NewContext(c.Request), g.UndoKey(c))
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

func index(prefix string) gin.HandlerFunc {
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

func newAction(prefix string) gin.HandlerFunc {
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

func create(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		dsClient, err := datastore.NewClient(c, "")
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		g := New(c, 0)
		withGame(c, g)

		err = g.FromForm(c, g.Type)
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

		ks, err := dsClient.AllocateIDs(c, []*datastore.Key{g.Header.Key})
		if err != nil {
			log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}
		g.Header.Key = ks[0]

		_, err = dsClient.RunInTransaction(c, func(tx *datastore.Transaction) error {
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

func accept(prefix string) gin.HandlerFunc {
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

		err = g.save(c)
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

func drop(prefix string) gin.HandlerFunc {
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
		err = g.save(c)

		if err != nil {
			log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
		}
		c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
	}
}

func fetch(c *gin.Context) {
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
		// pull from memcache/datastore
		// same as undo
		fallthrough
	case action == "undo":
		// pull from memcache/datastore
		if err := dsGet(c, g); err != nil {
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	default:
		if user.CurrentFrom(c) != nil {
			// pull from memcache and return if successful; otherwise pull from datastore
			if err := mcGet(c, g); err == nil {
				return
			}
		}

		log.Debugf("g: %#v", g)
		log.Debugf("k: %v", g.Header.Key)
		if err := dsGet(c, g); err != nil {
			log.Debugf("dsGet error: %v", err)
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}
	}
}

// pull temporary game state from memcache.  Note may be different from value stored in datastore.
func mcGet(c *gin.Context, g *Game) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	mkey := g.GetHeader().UndoKey(c)
	item, err := memcache.Get(appengine.NewContext(c.Request), mkey)
	if err != nil {
		return err
	}

	err = codec.Decode(g, item.Value)
	if err != nil {
		return err
	}

	err = g.afterCache()
	if err != nil {
		return err
	}

	withGame(c, g)
	color.WithMap(c, g.ColorMapFor(user.CurrentFrom(c)))
	return nil
}

// pull game state from memcache/datastore.  returned memcache should be same as datastore.
func dsGet(c *gin.Context, g *Game) error {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	dsClient, err := datastore.NewClient(c, "")
	if err != nil {
		return err
	}

	err = dsClient.Get(c, g.Header.Key, g.Header)
	switch {
	case err != nil:
		restful.AddErrorf(c, err.Error())
		return err
	case g == nil:
		err = fmt.Errorf("Unable to get game for id: %v", g.ID)
		restful.AddErrorf(c, err.Error())
		return err
	}

	s := newState()
	err = codec.Decode(&s, g.SavedState)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return err
	}

	g.State = s

	err = g.init(c)
	if err != nil {
		restful.AddErrorf(c, err.Error())
		return err
	}

	withGame(c, g)
	cm := g.ColorMapFor(user.CurrentFrom(c))
	color.WithMap(c, cm)
	return nil
}

func jsonIndexAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debugf("Entering")
		defer log.Debugf("Exiting")

		game.JSONIndexAction(c)
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
