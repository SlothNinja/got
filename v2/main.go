package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/patrickmn/go-cache"
)

const (
	NODE_ENV       = "NODE_ENV"
	production     = "production"
	userPrefix     = "user"
	gamesPrefix    = "games"
	ratingPrefix   = "rating"
	mailPrefix     = "mail"
	rootPath       = "/"
	hashKeyLength  = 64
	blockKeyLength = 32
	sessionName    = "sng-oauth"
)

func main() {
	if sn.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	db, err := datastore.NewClient(context.Background(), "")
	if err != nil {
		panic(fmt.Sprintf("unable to connect to database: %v", err.Error()))
	}

	// userClient := user.NewClient(db)

	mcache := cache.New(30*time.Minute, 10*time.Minute)

	s, err := getSecrets()
	if err != nil {
		panic(err.Error())
	}

	store := cookie.NewStore(s.HashKey, s.BlockKey)
	store.Options(sessions.Options{Domain: "slothninja.com"})
	// store := sessions.NewCookieStore([]byte("secret123"))

	r := gin.Default()
	// renderer := restful.ParseTemplates("templates/", ".tmpl")
	// r.HTMLRender = renderer

	r.Use(
		sessions.Sessions(sessionName, store),
	//	restful.AddTemplates(renderer.Templates),
	//	user.GetCUserHandler(userClient),
	)

	// Welcome Page (index.html) route
	// welcome.AddRoutes(r)

	// Games Routes
	// r = game.NewClient(db).AddRoutes(gamesPrefix, r)

	// User Routes
	// r = user_controller.NewClient(db).AddRoutes(userPrefix, r)

	// Rating Routes
	// r = rating.NewClient(db).AddRoutes(ratingPrefix, r)

	// After The Flood
	// r = atf.NewClient(db, mcache).Register(ATF, r)

	// Guild of Thieves
	r = NewClient(db, mcache).addRoutes(r)

	// Tammany Hall
	// r = tammany.NewClient(db, mcache).Register(Tammany, r)

	// Indonesia
	// r = indonesia.NewClient(db, mcache).Register(Indonesia, r)

	// Confucius
	// r = confucius.NewClient(db, mcache).Register(Confucius, r)

	// warmup
	r.GET("_ah/warmup", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r = staticRoutes(r)

	r.Run()
}

type secrets struct {
	HashKey   []byte
	BlockKey  []byte
	UpdatedAt time.Time
	Key       *datastore.Key `datastore:"__key__"`
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
func staticRoutes(r *gin.Engine) *gin.Engine {
	if sn.IsProduction() {
		return r
	}
	r.StaticFile("/", "dist/index.html")
	r.StaticFile("/app.js", "dist/app.js")
	r.StaticFile("/favicon.ico", "dist/favicon.ico")
	r.Static("/img", "dist/img")
	r.Static("/js", "dist/js")
	r.Static("/css", "dist/css")
	return r
}
