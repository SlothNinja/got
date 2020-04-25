module github.com/SlothNinja/got/v2

require (
	cloud.google.com/go/datastore v1.1.0
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/SlothNinja/codec v1.0.0
	github.com/SlothNinja/log v0.0.2
	github.com/SlothNinja/restful v1.0.0
	github.com/SlothNinja/schema v1.0.0
	github.com/SlothNinja/send v1.0.0
	github.com/SlothNinja/sn v1.0.0
	github.com/SlothNinja/user/v2 v2.0.0-alpha.6
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/gin v1.6.2
	github.com/gorilla/securecookie v1.1.1
	github.com/mailjet/mailjet-apiv3-go v0.0.0-20190724151621-55e56f74078c
	github.com/patrickmn/go-cache v2.1.0+incompatible
)

replace github.com/SlothNinja/sn => ./sn

go 1.13
