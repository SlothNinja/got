package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

const (
	production       = "production"
	userPrefix       = "user"
	gamesPrefix      = "games"
	ratingPrefix     = "rating"
	mailPrefix       = "mail"
	rootPath         = "/"
	hashKeyLength    = 64
	blockKeyLength   = 32
	sessionName      = "sng-oauth"
	NodeEnv          = "NODE_ENV"
	GotProjectIDEnv  = "GOT_PROJECT_ID"
	GotDSURLEnv      = "GOT_DS_URL"
	GotHostURLEnv    = "GOT_HOST_URL"
	UserProjectIDEnv = "USER_PROJECT_ID"
	UserDSURLEnv     = "USER_DS_URL"
	UserHostURLEnv   = "USER_HOST_URL"
)

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()

	if sn.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
		cl := newClient(ctx)
		defer cl.Close()
		cl.Router.Run()
	} else {
		gin.SetMode(gin.DebugMode)
		cl := newClient(ctx)
		defer cl.Close()
		cl.Router.RunTLS(getPort(), "cert.pem", "key.pem")
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

func (cl *client) staticRoutes() *client {
	if sn.IsProduction() {
		return cl
	}
	cl.Router.StaticFile("/", "dist/index.html")
	cl.Router.StaticFile("/index.html", "dist/index.html")
	cl.Router.StaticFile("/firebase-messaging-sw.js", "dist/firebase-messaging-sw.js")
	cl.Router.StaticFile("/manifest.json", "dist/manifest.json")
	cl.Router.StaticFile("/robots.txt", "dist/robots.txt")
	cl.Router.StaticFile("/precache-manifest.c0be88927a8120cb7373cf7df05f5688.js", "dist/precache-manifest.c0be88927a8120cb7373cf7df05f5688.js")
	cl.Router.StaticFile("/app.js", "dist/app.js")
	cl.Router.StaticFile("/favicon.ico", "dist/favicon.ico")
	cl.Router.Static("/img", "dist/img")
	cl.Router.Static("/js", "dist/js")
	cl.Router.Static("/css", "dist/css")
	return cl
}

func getGotProjectID() string {
	return os.Getenv(GotProjectIDEnv)
}

func getGotHostURL() string {
	return os.Getenv(GotHostURLEnv)
}

func getGotDSURL() string {
	return os.Getenv(GotDSURLEnv)
}

func getUserProjectID() string {
	return os.Getenv(UserProjectIDEnv)
}

func getUserDSURL() string {
	return os.Getenv(UserDSURLEnv)
}

func getUserHostURL() string {
	return os.Getenv(UserHostURLEnv)
}

func newLogClient() *log.Client {
	client, err := log.NewClient(getGotProjectID())
	if err != nil {
		log.Panicf("unable to create logging client: %v", err)
	}
	return client
}
