module go.osspkg.com/goppy/ormmysql

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
	github.com/go-sql-driver/mysql v1.8.1
	go.osspkg.com/goppy/errors v0.3.1
	go.osspkg.com/goppy/orm v0.3.5
	go.osspkg.com/goppy/plugins v0.3.1
	go.osspkg.com/goppy/routine v0.3.1
	go.osspkg.com/goppy/sqlcommon v0.3.3
	go.osspkg.com/goppy/xc v0.3.1
	go.osspkg.com/goppy/xlog v0.3.3
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	go.osspkg.com/goppy/iofile v0.3.3 // indirect
	go.osspkg.com/goppy/iosync v0.3.0 // indirect
	go.osspkg.com/goppy/syscall v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
