module go.osspkg.com/goppy

go 1.18

replace (
	go.osspkg.com/goppy/app => ./app
	go.osspkg.com/goppy/console => ./console
	go.osspkg.com/goppy/errors => ./errors
	go.osspkg.com/goppy/iosync => ./iosync
	go.osspkg.com/goppy/plugins => ./plugins
	go.osspkg.com/goppy/syscall => ./syscall
	go.osspkg.com/goppy/xc => ./xc
	go.osspkg.com/goppy/xlog => ./xlog
	go.osspkg.com/goppy/xtest => ./xtest
)

require (
	go.osspkg.com/goppy/app v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/console v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/errors v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/plugins v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	go.osspkg.com/algorithms v1.3.0 // indirect
	go.osspkg.com/goppy/iosync v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/syscall v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/xc v0.0.0-00010101000000-000000000000 // indirect
	go.osspkg.com/goppy/xlog v0.0.0-00010101000000-000000000000 // indirect
)
