module github.com/SlothNinja/got/v2

go 1.14

require (
	cloud.google.com/go/datastore v1.3.0
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/SlothNinja/color v1.0.0
	github.com/SlothNinja/contest v1.0.0
	github.com/SlothNinja/game v1.0.8
	github.com/SlothNinja/log v0.0.2
	github.com/SlothNinja/mlog v1.0.2
	github.com/SlothNinja/rating v1.0.5
	github.com/SlothNinja/restful v1.0.0
	github.com/SlothNinja/send v1.0.0
	github.com/SlothNinja/sn v1.0.1
	github.com/SlothNinja/type v1.0.1
	github.com/SlothNinja/undo v1.0.0
	github.com/SlothNinja/user v1.0.14
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/gin v1.6.3
	github.com/gorilla/securecookie v1.1.1
	github.com/hashicorp/go-retryablehttp v0.6.8
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/mailjet/mailjet-apiv3-go v0.0.0-20201009050126-c24bc15a9394
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	google.golang.org/api v0.36.0
	google.golang.org/grpc v1.34.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace github.com/SlothNinja/game => ./game
