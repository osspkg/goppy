module go.osspkg.com/goppy/geoip

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/ioutil => ../ioutil
	go.osspkg.com/goppy/plugins => ../plugins
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/web => ../web
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xnet => ../xnet
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	github.com/oschwald/geoip2-golang v1.9.0
	go.osspkg.com/goppy/plugins v0.3.1
	go.osspkg.com/goppy/web v0.3.4
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/oschwald/maxminddb-golang v1.12.0 // indirect
	go.osspkg.com/goppy/errors v0.3.1 // indirect
	go.osspkg.com/goppy/iosync v0.3.0 // indirect
	go.osspkg.com/goppy/ioutil v0.3.1 // indirect
	go.osspkg.com/goppy/syscall v0.3.0 // indirect
	go.osspkg.com/goppy/xc v0.3.1 // indirect
	go.osspkg.com/goppy/xlog v0.3.3 // indirect
	go.osspkg.com/goppy/xnet v0.3.1 // indirect
	go.osspkg.com/static v1.4.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
)
