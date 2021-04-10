package main

import (
	"context"
	"encoding/base64"
	"fmt"
	log2 "log"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/SlothNinja/cookie"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/rating"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"google.golang.org/api/option"
)

// type Client struct {
// 	*sn.Client
// }
//
// func NewClient(ctx context.Context, snClient *sn.Client) *Client {
// 	return &Client{snClient}
// }

type client struct {
	*sn.Client
	User      *user.Client
	MLog      *mlog.Client
	Rating    *rating.Client
	Messaging *messaging.Client
	logClient *log.Client
}

func newClient(ctx context.Context) *client {
	logClient := newLogClient()
	snClient := sn.NewClient(ctx, sn.Options{
		ProjectID: getGotProjectID(),
		DSURL:     getGotDSURL(),
		Logger:    logClient.Logger("got"),
		Cache:     cache.New(30*time.Minute, 10*time.Minute),
		Router:    gin.Default(),
	})

	uClient := user.NewClient(sn.NewClient(ctx, sn.Options{
		ProjectID: getUserProjectID(),
		DSURL:     getUserDSURL(),
		Logger:    snClient.Log,
		Cache:     snClient.Cache,
		Router:    snClient.Router,
	}))

	store, err := cookie.NewClient(uClient.Client).NewStore(ctx)
	if err != nil {
		snClient.Log.Panicf("unable create cookie store: %v", err)
	}
	snClient.Router.Use(sessions.Sessions(sessionName, store))

	nClient := &client{
		Client:    snClient,
		User:      uClient,
		MLog:      mlog.NewClient(snClient, uClient),
		Rating:    rating.NewClient(snClient, uClient, "rating"),
		Messaging: newMsgClient(ctx),
		logClient: logClient,
	}
	return nClient.addRoutes()
}

type CloseErrors struct {
	Client     error
	LogClient  error
	UserClient error
}

func (ce CloseErrors) Error() string {
	return fmt.Sprintf("error closing clients: client: %q logClient: %q userClient: %q",
		ce.Client, ce.LogClient, ce.UserClient)
}

func (cl *client) Close() error {
	var ce CloseErrors

	ce.Client = cl.Client.Close()
	ce.LogClient = cl.logClient.Close()
	ce.UserClient = cl.User.Client.Close()

	if ce.Client != nil || ce.LogClient != nil || ce.UserClient != nil {
		return ce
	}
	return nil
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

func (cl *client) login(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	referer := c.Request.Referer()
	encodedReferer := base64.StdEncoding.EncodeToString([]byte(referer))

	path := getUserHostURL() + "/login?redirect=" + encodedReferer
	cl.Log.Debugf("path: %q", path)
	c.Redirect(http.StatusSeeOther, path)
}

func (cl *client) logout(c *gin.Context) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	user.Logout(c)
	c.Redirect(http.StatusSeeOther, "/")
}
