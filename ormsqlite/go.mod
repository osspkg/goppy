module go.osspkg.com/goppy/ormsqlite

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iofile => ../iofile
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/orm => ../orm
	go.osspkg.com/goppy/plugins => ../plugins
	go.osspkg.com/goppy/routine => ../routine
	go.osspkg.com/goppy/sqlcommon => ../sqlcommon
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	github.com/mattn/go-sqlite3 v1.14.19
	go.osspkg.com/goppy/errors v0.2.0
	go.osspkg.com/goppy/iofile v0.2.0
	go.osspkg.com/goppy/orm v0.2.0
	go.osspkg.com/goppy/plugins v0.2.0
	go.osspkg.com/goppy/routine v0.2.0
	go.osspkg.com/goppy/sqlcommon v0.2.0
	go.osspkg.com/goppy/xc v0.2.0
	go.osspkg.com/goppy/xlog v0.2.0
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	go.osspkg.com/goppy/iosync v0.2.0 // indirect
	go.osspkg.com/goppy/syscall v0.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
