package main

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user/v2"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type Client struct {
	DS    *datastore.Client
	Game  sn.Client
	Cache *cache.Cache
}

func NewClient(dsClient *datastore.Client, mcache *cache.Cache) Client {
	return Client{
		DS:    dsClient,
		Game:  sn.NewClient(dsClient),
		Cache: mcache,
	}
}

func (client Client) addRoutes(engine *gin.Engine) *gin.Engine {
	// User Current
	engine.GET("/user/current", client.Current)

	// Game Group
	g := engine.Group("/game")

	// New
	g.GET("/new", client.newAction)

	// Create
	g.PUT("/new", client.create)

	// Show
	g.GET("/show/:hid",
		client.fetch,
		client.Game.GetMLog,
		client.show,
	)

	// Undo
	g.POST("/undo/:hid",
		client.fetch,
		client.undo,
	)

	// Finish
	g.POST("/finish/:hid",
		client.fetch,
		client.finish,
	)

	// Drop
	g.POST("/drop/:hid",
		client.fetch,
		client.drop,
	)

	// Accept
	g.POST("/accept/:hid",
		client.fetch,
		client.accept,
	)

	// Update
	g.PUT("/show/:hid",
		client.fetch,
		client.update,
	)

	// Add Message
	g.PUT("/show/:hid/addmessage",
		client.Game.GetMLog,
		client.Game.AddMessage(""),
	)

	// Games Group
	gs := engine.Group("/games")

	// Index
	// gs.GET("/:status", client.index)

	gs.GET("/:status/user/:uid", client.index)

	// JSON Data for Index
	gs.GET("/:status", client.jsonIndexAction)

	// JSON Data for Index
	gs.POST("/:status/user/:uid/json",
		client.jsonIndexAction,
	)

	// Admin Group
	admin := g.Group("/admin", user.RequireAdmin)

	admin.GET("/:hid",
		client.fetch,
		client.Game.GetMLog,
		client.show,
	)

	admin.POST("/:hid",
		client.fetch,
		client.update,
	)

	admin.PUT("/:hid",
		client.fetch,
		client.update,
	)

	return engine
}
