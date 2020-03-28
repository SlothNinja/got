package got

import (
	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/mlog"
	"github.com/SlothNinja/rating"
	"github.com/SlothNinja/user"
	stats "github.com/SlothNinja/user-stats"
	"github.com/gin-gonic/gin"
)

type Client struct {
	*datastore.Client
	Game   game.Client
	Rating rating.Client
	Stats  stats.Client
}

func NewClient(dsClient *datastore.Client) Client {
	return Client{
		Client: dsClient,
		Game:   game.NewClient(dsClient),
		Rating: rating.NewClient(dsClient),
		Stats:  stats.NewClient(dsClient),
	}
}

func (client Client) addRoutes(prefix string, engine *gin.Engine) *gin.Engine {
	/////////////////////////////////////////
	// Game Public Routes
	gpublic := engine.Group(prefix + "/game")

	// Show
	gpublic.GET("/show/:hid", client.show(prefix))

	///////////////////////////////////////////
	// Game Private Routes
	gprivate := engine.Group(prefix+"/game", user.RequireCurrentUser)

	// New
	gprivate.GET("/new", client.newAction(prefix))

	// Create
	gprivate.POST("", client.create(prefix))

	// Undo
	gprivate.POST("/undo/:hid", client.undo(prefix))

	// Finish
	gprivate.POST("/finish/:hid", client.finish(prefix))

	// Drop
	gprivate.POST("/drop/:hid", client.drop(prefix))

	// Accept
	gprivate.POST("/accept/:hid", client.accept(prefix))

	// Update
	gprivate.PUT("/show/:hid", client.update(prefix))

	// Add Message
	gprivate.PUT("/show/:hid/addmessage", mlog.AddMessage(prefix))

	///////////////////////////////////////////////////////////////
	// Games Public Routes
	gspublic := engine.Group(prefix + "/games")

	// Index
	gspublic.GET("/:status", client.index(prefix))

	gspublic.GET("/:status/user/:uid", client.index(prefix))

	// JSON Data for Index
	gspublic.POST("/:status/json", client.jsonIndexAction(prefix))

	// JSON Data for Index
	gspublic.POST("/:status/user/:uid/json", client.jsonIndexAction(prefix))

	////////////////////////////////////////////////////////////////
	// Game Admin Routes
	gadmin := engine.Group(prefix+"/game/admin", user.RequireAdmin)

	gadmin.GET("/:hid", client.show(prefix))

	gadmin.POST("/:hid", client.update(prefix))

	gadmin.PUT("/:hid", client.update(prefix))

	return engine
}
