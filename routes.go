package got

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/rating"
	"github.com/SlothNinja/sn"
	gtype "github.com/SlothNinja/type"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type Client struct {
	*sn.Client
	User    *user.Client
	MLog    *mlog.Client
	Rating  *rating.Client
	Game    *Game
	CUser   *user.User
	Prefix  string
	Context *gin.Context
}

func NewClient(dClient *datastore.Client, uClient *user.Client, mClient *mlog.Client, rClient *rating.Client,
	logger *log.Logger, cache *cache.Cache, router *gin.Engine, t gtype.Type) *Client {
	client := &Client{
		Client: sn.NewClient(dClient, logger, cache, router),
		User:   uClient,
		MLog:   mClient,
		Rating: rClient,
	}
	return client.register(t)
}

func (client *Client) addRoutes(prefix string) *Client {
	// Game Group
	g := client.Router.Group(prefix + "/game")

	// New
	g.GET("/new", client.newAction(prefix))

	// Create
	g.POST("", client.create(prefix))

	// Show
	g.GET("/show/:hid", client.show(prefix))

	// Undo
	g.POST("/undo/:hid", client.undo(prefix))

	// Finish
	g.POST("/finish/:hid", client.finish(prefix))

	// Drop
	g.POST("/drop/:hid", client.drop(prefix))

	// Accept
	g.POST("/accept/:hid", client.accept(prefix))

	// Update
	g.PUT("/show/:hid", client.update(prefix))

	// Add Message
	g.PUT("/show/:hid/addmessage", client.addMessage(prefix))

	// Games Group
	gs := client.Router.Group(prefix + "/games")

	// Index
	gs.GET("/:status", client.index(prefix))

	gs.GET("/:status/user/:uid", client.index(prefix))

	// JSON Data for Index
	gs.POST("/:status/json", client.jsonIndexAction(prefix))

	// JSON Data for Index
	gs.POST("/:status/user/:uid/json", client.jsonIndexAction(prefix))

	// Admin Group
	admin := g.Group("/admin")

	admin.GET("/:hid", client.show(prefix))

	admin.POST("/:hid", client.update(prefix))

	admin.PUT("/:hid", client.update(prefix))

	return client
}
