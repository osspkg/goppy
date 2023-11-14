module go.osspkg.com/goppy/app

go 1.18

replace (
	go.osspkg.com/goppy/console => ../console
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/algorithms v1.3.0
	go.osspkg.com/goppy/console v0.1.0
	go.osspkg.com/goppy/errors v0.1.0
	go.osspkg.com/goppy/syscall v0.1.0
	go.osspkg.com/goppy/xc v0.1.0
	go.osspkg.com/goppy/xlog v0.1.0
	go.osspkg.com/goppy/xtest v0.1.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	go.osspkg.com/goppy/iosync v0.1.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)
