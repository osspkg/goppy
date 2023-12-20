module go.osspkg.com/goppy/orm

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iofile => ../iofile
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/sqlcommon => ../sqlcommon
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.1.0
	go.osspkg.com/goppy/iofile v0.1.3
	go.osspkg.com/goppy/sqlcommon v0.1.4
	go.osspkg.com/goppy/xc v0.1.0
	go.osspkg.com/goppy/xlog v0.1.4
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	go.osspkg.com/goppy/iosync v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
