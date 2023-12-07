module go.osspkg.com/goppy/web

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/ioutil => ../ioutil
	go.osspkg.com/goppy/plugins => ../plugins
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xnet => ../xnet
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	github.com/mailru/easyjson v0.7.7
	go.osspkg.com/goppy/errors v0.1.0
	go.osspkg.com/goppy/iosync v0.1.2
	go.osspkg.com/goppy/ioutil v0.1.0
	go.osspkg.com/goppy/plugins v0.1.0
	go.osspkg.com/goppy/xc v0.1.0
	go.osspkg.com/goppy/xlog v0.1.3
	go.osspkg.com/goppy/xnet v0.1.1
	go.osspkg.com/goppy/xtest v0.1.2
	go.osspkg.com/static v1.4.0
)

require github.com/josharian/intern v1.0.0 // indirect
