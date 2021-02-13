module github.com/SlothNinja/got/v2

go 1.14

require (
	cloud.google.com/go/datastore v1.4.0
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/SlothNinja/color v1.0.0
	github.com/SlothNinja/contest v1.0.1
	github.com/SlothNinja/cookie v1.0.1
	github.com/SlothNinja/game v1.0.15
	github.com/SlothNinja/log v1.0.2
	github.com/SlothNinja/mlog v1.0.3
	github.com/SlothNinja/rating v1.0.7
	github.com/SlothNinja/restful v1.0.0
	github.com/SlothNinja/send v1.0.0
	github.com/SlothNinja/sn v1.0.3
	github.com/SlothNinja/type v1.0.1
	github.com/SlothNinja/undo v1.0.0
	github.com/SlothNinja/user v1.0.18
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/gin v1.6.3
	github.com/gorilla/securecookie v1.1.1
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/mailjet/mailjet-apiv3-go v0.0.0-20201009050126-c24bc15a9394
	github.com/patrickmn/go-cache v2.1.0+incompatible
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
)

replace github.com/SlothNinja/rating => ./private/rating

replace github.com/SlothNinja/contest => ./private/contest

replace github.com/SlothNinja/game => ./private/game

replace github.com/SlothNinja/send => ./private/send

replace github.com/SlothNinja/mlog => ./private/mlog
