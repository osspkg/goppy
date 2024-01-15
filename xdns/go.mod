module go.osspkg.com/goppy/xdns

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/plugins => ../plugins
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xnet => ../xnet
)

require (
	github.com/miekg/dns v1.1.57
	go.osspkg.com/goppy/errors v0.3.0
	go.osspkg.com/goppy/iosync v0.3.0
	go.osspkg.com/goppy/plugins v0.3.1
	go.osspkg.com/goppy/xc v0.3.0
	go.osspkg.com/goppy/xlog v0.3.0
	go.osspkg.com/goppy/xnet v0.3.0
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	go.osspkg.com/goppy/syscall v0.3.0 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
)
