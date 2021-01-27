package main

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type client struct {
	DS    *datastore.Client
	User  user.Client
	Cache *cache.Cache
}

func newClient(userClient user.Client, dsClient *datastore.Client, mcache *cache.Cache) client {
	return client{
		DS:    dsClient,
		User:  userClient,
		Cache: mcache,
	}
}

func (cl client) addRoutes(eng *gin.Engine) *gin.Engine {
	////////////////////////////////////////////
	// User Current
	eng.GET(cuPath, cl.current)

	////////////////////////////////////////////
	// Invitation Group
	inv := eng.Group(invitationPath)

	// New
	inv.GET(newPath, cl.newInvitation)

	// Create
	inv.PUT(newPath, cl.create)

	// Drop
	inv.PUT(dropPath, cl.drop)

	// Accept
	inv.PUT(acceptPath, cl.accept)

	// Details
	inv.GET(detailsPath, cl.details)

	/////////////////////////////////////////////
	// Invitations Group
	invs := eng.Group(invitationsPath)

	// Index
	invs.GET("", cl.invitationsIndex)

	/////////////////////////////////////////////
	// Game Group
	g := eng.Group(gamePath)

	// Show
	g.GET(showPath, cl.show)

	// Undo
	g.PUT(undoPath, cl.undo)

	// Redo
	g.PUT(redoPath, cl.redo)

	// Rest
	g.PUT(resetPath, cl.reset)

	// Place Thief Finish
	g.PUT(ptfinishPath, cl.placeThievesFinishTurn)

	// Move Thief Finish
	g.PUT(mtfinishPath, cl.moveThiefFinishTurn)

	// Passed Finish
	g.PUT(pfinishPath, cl.passedFinishTurn)

	// Place Thief
	g.PUT(placeThiefPath, cl.placeThief)

	// Play Card
	g.PUT(playCardPath, cl.playCard)

	// Select Thief
	g.PUT(selectThiefPath, cl.selectThief)

	// Move Thief
	g.PUT(moveThiefPath, cl.moveThief)

	// Pass
	g.PUT(passPath, cl.pass)

	// Message Group
	msg := g.Group(msgPath)

	// Add Message
	msg.PUT("add", cl.SN.AddMessage)

	// Get Message
	msg.GET("", cl.SN.GetMLog)

	// Games Group
	gs := eng.Group(gamesPath)

	// JSON Data for Index
	gs.GET(gamesIndexPath, cl.gamesIndex)

	// Admin Group
	admin := g.Group(adminPath, user.RequireAdmin)

	admin.GET(adminGetPath, cl.show)

	// Ratings
	eng = cl.SN.AddRoutes(ratingPrefix, eng)

	return eng
}
