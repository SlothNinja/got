package got

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/mlog"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

type server struct {
	*datastore.Client
}

func NewClient(dsClient *datastore.Client) server {
	return server{Client: dsClient}
}

func (s server) addRoutes(prefix string, engine *gin.Engine) *gin.Engine {
	// Game Group
	g := engine.Group(prefix + "/game")

	// New
	g.GET("/new",
		user.RequireCurrentUser(),
		gtype.SetTypes(),
		s.newAction(prefix),
	)

	// Create
	g.POST("",
		user.RequireCurrentUser(),
		s.create(prefix),
	)

	// Show
	g.GET("/show/:hid",
		s.fetch,
		mlog.Get,
		game.SetAdmin(false),
		s.show(prefix),
	)

	// Undo
	g.POST("/undo/:hid",
		s.fetch,
		s.undo(prefix),
	)

	// Finish
	g.POST("/finish/:hid",
		s.fetch,
		stats.Fetch(user.CurrentFrom),
		s.finish(prefix),
	)

	// Drop
	g.POST("/drop/:hid",
		user.RequireCurrentUser(),
		s.fetch,
		s.drop(prefix),
	)

	// Accept
	g.POST("/accept/:hid",
		user.RequireCurrentUser(),
		s.fetch,
		s.accept(prefix),
	)

	// Update
	g.PUT("/show/:hid",
		user.RequireCurrentUser(),
		s.fetch,
		game.RequireCurrentPlayerOrAdmin(),
		game.SetAdmin(false),
		s.update(prefix),
	)

	// Add Message
	g.PUT("/show/:hid/addmessage",
		user.RequireCurrentUser(),
		mlog.Get,
		mlog.AddMessage(prefix),
	)

	// Games Group
	gs := engine.Group(prefix + "/games")

	// Index
	gs.GET("/:status",
		gtype.SetTypes(),
		s.index(prefix),
	)

	gs.GET("/:status/user/:uid",
		gtype.SetTypes(),
		s.index(prefix),
	)

	// JSON Data for Index
	gs.POST("/:status/json",
		gtype.SetTypes(),
		game.GetFiltered(gtype.GOT),
		s.jsonIndexAction(prefix),
	)

	// JSON Data for Index
	gs.POST("/:status/user/:uid/json",
		gtype.SetTypes(),
		game.GetFiltered(gtype.GOT),
		s.jsonIndexAction(prefix),
	)

	// Admin Group
	admin := g.Group("/admin", user.RequireAdmin)

	admin.GET("/:hid",
		s.fetch,
		mlog.Get,
		game.SetAdmin(true),
		s.show(prefix),
	)

	admin.POST("/admin/:hid",
		user.RequireCurrentUser(),
		s.fetch,
		game.RequireCurrentPlayerOrAdmin(),
		game.SetAdmin(true),
		s.update(prefix),
	)

	admin.PUT("/admin/:hid",
		user.RequireCurrentUser(),
		s.fetch,
		game.RequireCurrentPlayerOrAdmin(),
		game.SetAdmin(true),
		s.update(prefix),
	)

	return engine
}
