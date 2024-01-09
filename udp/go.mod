module go.osspkg.com/goppy/udp

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xnet => ../xnet
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/iosync v0.1.5
	go.osspkg.com/goppy/xc v0.1.1
	go.osspkg.com/goppy/xlog v0.1.7
	go.osspkg.com/goppy/xnet v0.1.2
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	go.osspkg.com/goppy/errors v0.1.1 // indirect
	go.osspkg.com/goppy/syscall v0.1.2 // indirect
)
