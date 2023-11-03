module go.osspkg.com/goppy/auth

go 1.18

replace (
	go.osspkg.com/goppy/encryption => ../encryption
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/ioutil => ../ioutil
	go.osspkg.com/goppy/plugins => ../plugins
	go.osspkg.com/goppy/random => ../random
	go.osspkg.com/goppy/web => ../web
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xnet => ../xnet
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	github.com/mailru/easyjson v0.7.7
	go.osspkg.com/goppy/encryption v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/errors v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/ioutil v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/plugins v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/random v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/web v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/xtest v0.0.0-00010101000000-000000000000
	golang.org/x/oauth2 v0.13.0
)

require (
	cloud.google.com/go/compute v1.20.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	go.osspkg.com/goppy/iosync v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/xc v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/xlog v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/xnet v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/static v1.4.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
