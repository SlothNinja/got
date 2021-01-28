package got

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/codec"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/sn"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
)

const (
	gameKey   = "Game"
	homePath  = "/"
	jsonKey   = "JSON"
	statusKey = "Status"
	msgEnter  = "Entering"
	msgExit   = "Exiting"
)

var (
	ErrInvalidID    = errors.New("invalid identifier")
	ErrNotFound     = errors.New("not found")
	ErrInvalidCache = errors.New("invalid cache value")
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

func (client *Client) show(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c

		id, err := getID(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		ml, err := client.MLog.Get(c, id)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		client.CUser, err = client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
		}

		err = client.getGameFor(id)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		c.HTML(http.StatusOK, prefix+"/show", gin.H{
			"CUser":      client.CUser,
			"Game":       client.Game,
			"MessageLog": ml,
			"ColorMap":   client.Game.ColorMapFor(client.CUser),
		})
	}
}

func (client *Client) addMessage(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c
		g, cu := client.Game, client.CUser

		id, err := getID(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		cu, err = client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
			return
		}

		err = client.getGameFor(id)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		ml, err := client.MLog.Get(c, id)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		m := ml.AddMessage(cu, c.PostForm("message"))

		_, err = client.MLog.Put(c, id, ml)
		if err != nil {
			client.Log.Errorf(err.Error())
			return
		}

		c.HTML(http.StatusOK, "shared/message", gin.H{
			"message": m,
			"ctx":     c,
			"map":     g.ColorMapFor(cu),
			"link":    cu.Link(),
		})
	}
}

func (client *Client) dispatch(c *gin.Context) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	a := c.PostForm("action")
	switch a {
	case "select-area":
		client.selectArea(c)
	case "admin-header":
		client.Game.adminHeader(c, client.CUser)
	case "admin-player":
		client.adminPlayer()
	case "pass":
		client.pass()
	default:
		client.Log.Errorf("%v is not a valid action", a)
	}
}

func (client *Client) undo(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c
		g, cu := client.Game, client.CUser

		path := showPath(prefix, c.Param("hid"))

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		id, err := getID(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		err = client.dsGet(id)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		client.Cache.Delete(g.UndoKey(cu))
		client.Cache.Delete(g.Key.Encode())
		c.Redirect(http.StatusSeeOther, path)
	}
}

func (client *Client) update(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c

		var err error

		client.CUser, err = client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}

		id, err := getID(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}

		err = client.getGameFor(id)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, homePath)
			return
		}

		client.dispatch(c)
	}
}

func (client *Client) flashError(err error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	client.Context.HTML(http.StatusOK, "got/flash_notice", gin.H{
		"CUser":  client.CUser,
		"Errors": []template.HTML{template.HTML(err.Error())},
	})
}

func (client *Client) html(tmpl string) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	client.Context.HTML(http.StatusOK, tmpl, gin.H{
		"CUser":   client.CUser,
		"Game":    client.Game,
		"Notices": restful.NoticesFrom(client.Context),
		"Errors":  restful.ErrorsFrom(client.Context),
	})
}

func (client *Client) save() error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	g, c, cu := client.Game, client.Context, client.CUser

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

	client.Cache.Delete(g.UndoKey(cu))
	return nil
}

func (client *Client) saveWith(c *gin.Context, g *Game, cu *user.User, s *user.Stats, cs []*contest.Contest) error {
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

		_, err = client.User.StatsUpdate(c, s, g.UpdatedAt)
		if err != nil {
			return err
		}

		l := len(cs) + 1
		ks := make([]*datastore.Key, l)
		es := make([]interface{}, l)
		ks[0], es[0] = g.Key, g.Header
		for i, contest := range cs {
			ks[i+1], es[i+1] = contest.Key, contest
		}

		_, err = tx.PutMulti(ks, es)
		if err != nil {
			return err
		}

		client.Cache.Delete(g.UndoKey(cu))
		return nil
	})
	return err
}

func wrap(s *user.Stats, cs []*contest.Contest) ([]*datastore.Key, []interface{}) {
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
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

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

func (client *Client) index(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c

		gs := game.GamersFrom(c)
		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Debugf(err.Error())
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
//		log.Debugf(ctx, msgEnter)
//		defer log.Debugf(ctx, msgExit)
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

func runningPath(prefix string) string {
	return fmt.Sprintf("/%s/games/running", prefix)
}

func newPath(prefix string) string {
	return fmt.Sprintf("/%s/game/new", prefix)
}

func (client *Client) newAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c

		g := New(c, 0)
		withGame(c, g)
		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.FromParams(c, cu, gtype.GOT)
		if err != nil {
			client.Log.Errorf(err.Error())
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

func (client *Client) create(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c

		g := New(c, 0)
		withGame(c, g)

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.FromForm(c, cu, g.Type)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.fromForm(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		err = g.encode(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		client.Log.Debugf("before client.DS.AllocateIDs")
		ks, err := client.DS.AllocateIDs(c, []*datastore.Key{g.Header.Key})
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}
		g.Header.Key = ks[0]

		client.Log.Debugf("before client.DS.RunInTransaction")
		_, err = client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
			m := mlog.New(g.Header.Key.ID)
			ks := []*datastore.Key{g.Header.Key, m.Key}
			es := []interface{}{g.Header, m}

			_, err := tx.PutMulti(ks, es)
			return err
		})
		if err != nil {
			client.Log.Errorf(err.Error())
			c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
			return
		}

		restful.AddNoticef(c, "<div>%s created.</div>", g.Title)
		c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
		return
	}
}

func (client *Client) accept(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c
		g, cu := client.Game, client.CUser

		path := recruitingPath(prefix)

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		id, err := getID(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		err = client.getGameFor(id)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		start, err := g.Accept(c, cu)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		if start {
			err = g.Start(c)
			if err != nil {
				client.Log.Errorf(err.Error())
				restful.AddErrorf(c, err.Error())
				c.Redirect(http.StatusSeeOther, path)
				return
			}
		}

		err = client.save()
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
			return
		}

		if start {
			err = g.SendTurnNotificationsTo(c, g.CurrentPlayer())
			if err != nil {
				restful.AddErrorf(c, err.Error())
				client.Log.Errorf(err.Error())
			}
		}
		c.Redirect(http.StatusSeeOther, recruitingPath(prefix))
	}
}

func (client *Client) drop(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c
		g, cu := client.Game, client.CUser

		path := recruitingPath(prefix)

		cu, err := client.User.Current(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
		}

		id, err := getID(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
		}

		err = client.getGameFor(id)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
		}

		err = g.Drop(cu)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
		}

		err = client.save()
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
		}

		c.Redirect(http.StatusSeeOther, path)
	}
}

func (client *Client) jsonIndexAction(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client.Log.Debugf(msgEnter)
		defer client.Log.Debugf(msgExit)

		client.Prefix, client.Context = prefix, c
		cu := client.CUser

		path := runningPath(prefix)

		var err error
		client.CUser, err = client.User.Current(c)
		if err != nil {
			c.JSON(http.StatusOK, fmt.Sprintf("%v", err))
			return
		}

		gs, cnt, err := client.getFiltered(c)
		if err != nil {
			client.Log.Errorf(err.Error())
			restful.AddErrorf(c, err.Error())
			c.Redirect(http.StatusSeeOther, path)
		}

		grs := make([]game.Gamer, len(gs))
		for i, g := range gs {
			grs[i] = g
		}

		data, err := game.ToGameTable(c, grs, cnt, cu)
		if err != nil {
			c.JSON(http.StatusOK, fmt.Sprintf("%v", err))
			return
		}
		c.JSON(http.StatusOK, data)
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

func getID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("hid"), 10, 64)
	if err != nil {
		return -1, ErrInvalidID
	}
	return id, nil
}

func (client *Client) getGameFor(id int64) error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	g := client.Game

	err := client.mcGetFor(id)
	if err == nil {
		return nil
	}

	err = client.dsGet(id)
	if err != nil {
		return err
	}
	client.Cache.SetDefault(g.Key.Encode(), g)
	return nil
}

// pull temporary game state from cache.  Note may be different from value stored in datastore.
func (client *Client) mcGetFor(id int64) error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	g, cu, c := client.Game, client.CUser, client.Context
	err := client.mcGet(id)

	switch {
	case cu == nil:
		return err
	case err == nil && !g.IsCurrentPlayer(cu):
		return nil
	}

	g1 := New(c, id)
	mkey := g1.UndoKey(cu)
	item, found := client.Cache.Get(mkey)
	if !found {
		if err == nil {
			return nil
		}
		return ErrNotFound
	}

	g2, ok := item.(*Game)
	if !ok {
		client.Cache.Delete(mkey)
		if err == nil {
			return nil
		}
		return ErrInvalidCache
	}
	g2.SetCTX(c)

	client.Game = g2
	return nil
}

// pull temporary game state from cache.  Note may be different from value stored in datastore.
func (client *Client) mcGet(id int64) error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	c := client.Context
	g := New(c, id)
	k := g.Key.Encode()
	item, found := client.Cache.Get(k)
	if !found {
		return ErrNotFound
	}

	g2, ok := item.(*Game)
	if !ok {
		client.Cache.Delete(k)
		return ErrInvalidCache
	}
	g2.SetCTX(c)

	client.Game = g2
	return nil
}

// pull game state from cache/datastore.  returned cache should be same as datastore.
func (client *Client) dsGet(id int64) error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	c := client.Context
	g := New(c, id)
	err := client.DS.Get(c, g.Key, g.Header)
	if err != nil {
		return err
	}

	state := newState()
	err = codec.Decode(&state, g.SavedState)
	if err != nil {
		return err
	}

	g.State = state

	g.init()
	client.Game = g
	return nil
}

func getAllQuery(c *gin.Context) *datastore.Query {
	return datastore.NewQuery("Game").Ancestor(game.GamesRoot(c))
}

func (client *Client) getFiltered(c *gin.Context) ([]*Game, int64, error) {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	status, sid, start, length := c.Param("status"), c.Param("uid"), c.PostForm("start"), c.PostForm("length")
	q := getAllQuery(c).KeysOnly()

	if status != "" {
		st := game.ToStatus[strings.ToLower(status)]
		q = q.Filter("Status=", int(st))
	}

	if sid != "" {
		id, err := strconv.Atoi(sid)
		if err == nil {
			q = q.Filter("UserIDS=", id)
		}
	}

	q = q.Filter("Type=", int(gtype.GOT)).Order("-UpdatedAt")

	cnt, err := client.DS.Count(c, q)
	if err != nil {
		return nil, 0, err
	}

	if start != "" {
		st, err := strconv.ParseInt(start, 10, 32)
		if err == nil {
			q = q.Offset(int(st))
		}
	}

	if length != "" {
		l, err := strconv.ParseInt(length, 10, 32)
		if err == nil {
			q = q.Limit(int(l))
		}
	}

	ks, err := client.DS.GetAll(c, q, nil)
	if err != nil {
		return nil, 0, err
	}

	gs := make([]*Game, len(ks))
	me := make(datastore.MultiError, len(ks))
	isNil := true
	for i, k := range ks {
		err = client.getGameFor(k.ID)
		if err != nil {
			isNil = false
		}
		g := *(client.Game)
		gs[i] = &g
	}
	if isNil {
		return gs, int64(cnt), nil
	}
	return gs, int64(cnt), me
}

// func (client *Client) cache(g *Game) {
// 	client.Cache.SetDefault(g.Key.Encode(), g)
// }
//
// func (client *Client) cacheFor(g *Game, u *user.User) {
// 	if u == nil {
// 		client.cache(g)
// 		return
// 	}
// 	k := g.UndoKey(u)
// 	client.Cache.SetDefault(k, g)
// }
//
// func (client *Client) uncache() {
// 	client.Cache.Delete(client.Game.Key.Encode())
// }
//
// func (client *Client) uncacheFor(u *user.User) {
// 	client.uncache()
// 	if u == nil {
// 		return
// 	}
// 	k := client.Game.UndoKey(u)
// 	client.Cache.Delete(k)
// }
