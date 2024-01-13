module go.osspkg.com/goppy/sqlcommon

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.1.2
	go.osspkg.com/goppy/xlog v0.1.7
	go.osspkg.com/goppy/xtest v0.1.4
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	go.osspkg.com/goppy/iosync v0.1.5 // indirect
	go.osspkg.com/goppy/syscall v0.1.3 // indirect
)
