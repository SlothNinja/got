package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/cookie"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/SlothNinja/user"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/patrickmn/go-cache"
)

const (
	nodeEnv            = "NODE_ENV"
	production         = "production"
	userPrefix         = "user"
	gamesPrefix        = "games"
	ratingPrefix       = "rating"
	mailPrefix         = "mail"
	rootPath           = "/"
	hashKeyLength      = 64
	blockKeyLength     = 32
	sessionName        = "sng-oauth"
	LOGIN_HOST         = "LOGIN_HOST"
	googleCloudProject = "GOOGLE_CLOUD_PROJECT"
)

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	if sn.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	ctx := context.Background()
	db, err := datastore.NewClient(ctx, "")
	if err != nil {
		panic(fmt.Sprintf("unable to connect to database: %v", err.Error()))
	}

	// userClient := user.NewClient(db)

	mcache := cache.New(30*time.Minute, 10*time.Minute)

	// s, err := getSecrets()
	// if err != nil {
	// 	panic(err.Error())
	// }

	// store := cookie.NewStore(s.HashKey, s.BlockKey)
	// store.Options(sessions.Options{Domain: "slothninja.com"})
	// // store := sessions.NewCookieStore([]byte("secret123"))
	logClient := newLogClient()
	defer logClient.Close()

	logger := logClient.Logger("got")
	store, err := cookie.NewClient(logger, mcache).NewStore()
	if err != nil {
		logger.Panicf("unable create cookie store: %v", err)
	}

	router := gin.Default()
	// renderer := restful.ParseTemplates("templates/", ".tmpl")
	// r.HTMLRender = renderer

	router.Use(sessions.Sessions(sessionName, store))

	userClient := user.NewClient(logger, mcache)

	// Guild of Thieves
	cl := newClient(ctx, db, userClient, logger, mcache, router)

	if gin.Mode() == gin.DebugMode {
		cl.Router.RunTLS(getPort(), "cert.pem", "key.pem")
	} else {
		cl.Router.Run()
	}
}

type secrets struct {
	HashKey   []byte
	BlockKey  []byte
	UpdatedAt time.Time
	Key       *datastore.Key `datastore:"__key__"`
}

func getPort() string {
	return ":" + os.Getenv("PORT")
}

func getSecrets() (secrets, error) {
	hashKey, err := base64.StdEncoding.DecodeString("v9UGh93EVzBPzfezwYCsZzfuL1LzaP8KVD4fAidyL1UmnsMqL5cnOQanWa7nE/tb3eBmUyv4ci66K+rnDs6CGA==")
	if err != nil {
		return secrets{}, err
	}
	blockKey, err := base64.StdEncoding.DecodeString("DT0/WyGLqwBYuo/l82Gq1DCxq/sVhVrTuzMFRJxPDQU=")
	if err != nil {
		return secrets{}, err
	}
	return secrets{
		HashKey:  hashKey,
		BlockKey: blockKey,
	}, nil
}

func secretsKey() *datastore.Key {
	return datastore.NameKey("Secrets", "root", nil)
}

func genSecrets() (secrets, error) {
	s := secrets{
		HashKey:  securecookie.GenerateRandomKey(hashKeyLength),
		BlockKey: securecookie.GenerateRandomKey(blockKeyLength),
		Key:      secretsKey(),
	}

	if s.HashKey == nil {
		return s, fmt.Errorf("generated hashKey was nil")
	}

	if s.BlockKey == nil {
		return s, fmt.Errorf("generated blockKey was nil")
	}

	return s, nil
}

func (s *secrets) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(s, ps)
}

func (s *secrets) Save() ([]datastore.Property, error) {
	s.UpdatedAt = time.Now()
	return datastore.SaveStruct(s)
}

func (s *secrets) LoadKey(k *datastore.Key) error {
	s.Key = k
	return nil
}

// staticHandler for local development since app.yaml is ignored
// static files are handled via app.yaml routes when deployed
// func staticRoutes(r *gin.Engine) *gin.Engine {
// 	if sn.IsProduction() {
// 		return r
// 	}
// 	r.StaticFile("/", "dist/index.html")
// 	r.StaticFile("/app.js", "dist/app.js")
// 	r.StaticFile("/favicon.ico", "dist/favicon.ico")
// 	r.Static("/img", "dist/img")
// 	r.Static("/js", "dist/js")
// 	r.Static("/css", "dist/css")
// 	return r
// }

func (cl *client) staticRoutes() *client {
	if sn.IsProduction() {
		return cl
	}
	// cl.Router.StaticFile("/favicon.ico", "public/favicon.ico")
	// cl.Router.Static("/images", "public/images")
	// cl.Router.Static("/javascripts", "public/javascripts")
	// cl.Router.Static("/js", "public/js")
	// cl.Router.Static("/stylesheets", "public/stylesheets")
	// cl.Router.Static("/rules", "public/rules")
	// cl.Router.StaticFile("/", "dist/index.html")
	cl.Router.StaticFile("/", "dist/index.html")
	cl.Router.StaticFile("/index.html", "dist/index.html")
	cl.Router.StaticFile("/firebase-messaging-sw.js", "dist/firebase-messaging-sw.js")
	cl.Router.StaticFile("/manifest.json", "dist/manifest.json")
	cl.Router.StaticFile("/robots.txt", "dist/robots.txt")
	cl.Router.StaticFile("/precache-manifest.169265351c010cfd724fe63afe555e11.js", "dist/precache-manifest.169265351c010cfd724fe63afe555e11.js")
	// cl.Router.StaticFile("/firebase-messaging-sw.js", "dist/firebase-messaging-sw.js")
	cl.Router.StaticFile("/app.js", "dist/app.js")
	cl.Router.StaticFile("/favicon.ico", "dist/favicon.ico")
	cl.Router.Static("/img", "dist/img")
	cl.Router.Static("/js", "dist/js")
	cl.Router.Static("/css", "dist/css")
	return cl
}

func getProjectID() string {
	return os.Getenv(googleCloudProject)
}

func newLogClient() *log.Client {
	client, err := log.NewClient(getProjectID())
	if err != nil {
		log.Panicf("unable to create logging client: %v", err)
	}
	return client
}
