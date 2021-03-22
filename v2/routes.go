package main

import (
	"context"
	"encoding/base64"
	log2 "log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/rating"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"google.golang.org/api/option"
)

type client struct {
	*sn.Client
	User      *user.Client
	MLog      *mlog.Client
	Rating    *rating.Client
	Messaging *messaging.Client
}

func newClient(ctx context.Context, dClient *datastore.Client, uClient *user.Client, logger *log.Logger, cache *cache.Cache, router *gin.Engine) *client {
	cl := &client{
		Client:    sn.NewClient(dClient, logger, cache, router),
		User:      uClient,
		MLog:      mlog.NewClient(dClient, uClient, logger, cache),
		Rating:    rating.NewClient(dClient, uClient, logger, cache, router, "rating"),
		Messaging: newMsgClient(ctx),
	}
	return cl.addRoutes()
}

const GotCreds = "GOT_CREDS"

func newMsgClient(ctx context.Context) *messaging.Client {
	if sn.IsProduction() {
		log.Debugf("production")
		app, err := firebase.NewApp(ctx, nil)
		if err != nil {
			log2.Panicf("unable to create messaging client: %v", err)
			return nil
		}
		cl, err := app.Messaging(ctx)
		if err != nil {
			log2.Panicf("unable to create messaging client: %v", err)
			return nil
		}
		return cl
	}
	log.Debugf("development")
	app, err := firebase.NewApp(
		ctx,
		nil,
		option.WithGRPCConnectionPool(50),
		option.WithCredentialsFile(os.Getenv(GotCreds)),
	)
	if err != nil {
		log2.Panicf("unable to create messaging client: %v", err)
		return nil
	}
	cl, err := app.Messaging(ctx)
	if err != nil {
		log2.Panicf("unable to create messaging client: %v", err)
		return nil
	}
	return cl
}

func (cl *client) addRoutes() *client {
	cl.staticRoutes()

	// warmup
	cl.Router.GET("_ah/warmup", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// login
	cl.Router.GET("login", cl.login)

	// login
	cl.Router.GET("logout", cl.logout)

	////////////////////////////////////////////
	// Home
	cl.Router.GET(homePath, cl.homeHandler)

	////////////////////////////////////////////
	// Message Log
	msg := cl.Router.Group("mlog")

	// Get
	msg.GET("/:id", cl.mlogHandler)

	// Add
	msg.PUT("/:id/add", cl.mlogAddHandler)

	////////////////////////////////////////////
	// Invitation Group
	inv := cl.Router.Group(invitationPath)

	// New
	inv.GET(newPath, cl.newInvitationHandler)

	// Create
	inv.PUT(newPath, cl.createHandler)

	// Drop
	inv.PUT(dropPath, cl.dropHandler)

	// Accept
	inv.PUT(acceptPath, cl.acceptHandler)

	// Details
	inv.GET(detailsPath, cl.details)

	/////////////////////////////////////////////
	// Invitations Group
	invs := cl.Router.Group(invitationsPath)

	// Index
	invs.POST("", cl.invitationsIndexHandler)

	/////////////////////////////////////////////
	// Game Group
	g := cl.Router.Group(gamePath)

	// Show
	g.GET(showPath, cl.showHandler)

	// Subscribe
	g.PUT(subscribePath, cl.subscribeHandler)

	// Unsubscribe
	g.PUT(unsubscribePath, cl.unsubscribeHandler)

	// Undo
	g.PUT(undoPath, cl.undo)

	// Redo
	g.PUT(redoPath, cl.redo)

	// Reset
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
	gs := cl.Router.Group(gamesPath)

	// JSON Data for Index
	gs.POST(gamesIndexPath, cl.gamesIndex)

	// Admin Group
	// admin := g.Group(adminPath, user.RequireAdmin)

	// admin.GET(adminGetPath, cl.show)

	// Ratings
	// eng = cl.SN.AddRoutes(ratingPrefix, eng)
	return cl
}

func getLoginHost() string {
	return os.Getenv(LOGIN_HOST)
}

func (cl *client) login(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	referer := c.Request.Referer()
	encodedReferer := base64.StdEncoding.EncodeToString([]byte(referer))

	c.Redirect(http.StatusSeeOther, getLoginHost()+"/login?redirect="+encodedReferer)
}

func (cl *client) logout(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	user.Logout(c)
	c.Redirect(http.StatusSeeOther, "/")
}
