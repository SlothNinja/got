package main

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/sn/v2"
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
	////////////////////////////////////////////
	// User Current
	engine.GET(cuPath, client.Current)

	////////////////////////////////////////////
	// Invitation Group
	inv := engine.Group(invitationPath)

	// New
	inv.GET(newPath, client.newInvitation)

	// Create
	inv.PUT(newPath, client.create)

	// Drop
	inv.PUT(dropPath, client.drop)

	// Accept
	inv.PUT(acceptPath, client.accept)

	/////////////////////////////////////////////
	// Invitations Group
	invs := engine.Group(invitationsPath)

	// Index
	invs.GET("", client.invitationsIndex)

	// Game Group
	g := engine.Group(gamePath)

	// Show
	g.GET(showPath, client.show)

	// Undo
	g.PUT(undoPath, client.undo)

	// Redo
	g.PUT(redoPath, client.redo)

	// Rest
	g.PUT(resetPath, client.reset)

	// Finish
	g.POST(finishPath, client.finish)

	// Update
	g.PUT(updatePath, client.update)

	// Update
	g.PUT(placeThiefPath, client.placeThief)

	// Add Message
	g.PUT(msgPath, client.Game.AddMessage(""))

	// Games Group
	gs := engine.Group(gamesPath)

	// Index
	// gs.GET("/:status", client.index)

	gs.GET(indexPath, client.index)

	// JSON Data for Index
	gs.GET(gamesIndexPath, client.gamesIndex)

	// // JSON Data for Index
	// gs.POST("/:status/user/:uid/json",
	// 	client.jsonIndexAction,
	// )

	// Admin Group
	admin := g.Group(adminPath, user.RequireAdmin)

	admin.GET(adminGetPath, client.show)

	admin.PUT(adminPutPath, client.update)

	return engine
}
