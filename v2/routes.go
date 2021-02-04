package main

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/rating"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// type client struct {
// 	DS    *datastore.Client
// 	SN    sn.Client
// 	Cache *cache.Cache
// }
//
// func newClient(dsClient *datastore.Client, mcache *cache.Cache) client {
// 	return client{
// 		DS:    dsClient,
// 		SN:    sn.NewClient(dsClient),
// 		Cache: mcache,
// 	}
// }

type client struct {
	*sn.Client
	User   *user.Client
	MLog   *mlog.Client
	Rating *rating.Client
	g      *Game
	gc     *gcommitted
	cu     *user.User
	cp     *player
	ctx    *gin.Context
}

func (cl *client) CUser() *user.User {
	if cl.cu != nil {
		return cl.cu
	}

	var err error
	cl.cu, err = cl.User.Current(cl.ctx)
	if err != nil {
		cl.Log.Errorf(err.Error())
		cl.cu = nil
		return nil
	}
	return cl.cu
}

func newClient(dClient *datastore.Client, uClient *user.Client, logger *log.Logger, cache *cache.Cache, router *gin.Engine) *client {
	cl := &client{
		Client: sn.NewClient(dClient, logger, cache, router),
		User:   uClient,
		MLog:   mlog.NewClient(dClient, uClient, logger, cache),
		Rating: rating.NewClient(dClient, uClient, logger, cache, router, "rating"),
	}
	return cl.staticRoutes()
}

func (cl *client) addRoutes(eng *gin.Engine) *gin.Engine {
	////////////////////////////////////////////
	// User Current
	eng.GET(cuPath, cl.cuHandler)

	////////////////////////////////////////////
	// Invitation Group
	inv := eng.Group(invitationPath)

	// New
	inv.GET(newPath, cl.newInvitationHandler)

	// Create
	inv.PUT(newPath, cl.createHandler)

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
	g.PUT(ptfinishPath, cl.placeThievesFinishTurnHandler)

	// Move Thief Finish
	g.PUT(mtfinishPath, cl.moveThiefFinishTurnHandler)

	// Passed Finish
	g.PUT(pfinishPath, cl.passedFinishTurnHandler)

	// Place Thief
	g.PUT(placeThiefPath, cl.placeThiefHandler)

	// Play Card
	g.PUT(playCardPath, cl.playCardHandler)

	// Select Thief
	g.PUT(selectThiefPath, cl.selectThiefHandler)

	// Move Thief
	g.PUT(moveThiefPath, cl.moveThiefHandler)

	// Pass
	g.PUT(passPath, cl.passHandler)

	// Games Group
	gs := eng.Group(gamesPath)

	// JSON Data for Index
	gs.GET(gamesIndexPath, cl.gamesIndex)

	// Admin Group
	// admin := g.Group(adminPath, user.RequireAdmin)

	// admin.GET(adminGetPath, cl.show)

	// Ratings
	// eng = cl.SN.AddRoutes(ratingPrefix, eng)

	return eng
}
