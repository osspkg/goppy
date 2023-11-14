module go.osspkg.com/goppy/ws

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	github.com/gorilla/websocket v1.5.0
	github.com/mailru/easyjson v0.7.7
	go.osspkg.com/goppy v0.15.0
	go.osspkg.com/goppy/errors v0.1.0
	go.osspkg.com/goppy/iosync v0.1.0
	go.osspkg.com/goppy/xc v0.1.0
	go.osspkg.com/goppy/xlog v0.1.0
	go.osspkg.com/goppy/xtest v0.1.0
)

require github.com/josharian/intern v1.0.0 // indirect