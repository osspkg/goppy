module go.osspkg.com/goppy/geoip

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/ioutil => ../ioutil
	go.osspkg.com/goppy/plugins => ../plugins
	go.osspkg.com/goppy/web => ../web
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xnet => ../xnet
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	github.com/oschwald/geoip2-golang v1.9.0
	go.osspkg.com/goppy/plugins v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/web v0.0.0-00010101000000-000000000000
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/oschwald/maxminddb-golang v1.11.0 // indirect
	go.osspkg.com/goppy/errors v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/iosync v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/ioutil v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/xc v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/xlog v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/xnet v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/static v1.4.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
