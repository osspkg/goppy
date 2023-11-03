module go.osspkg.com/goppy/xlog

go 1.18

replace (
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	github.com/mailru/easyjson v0.7.7
	go.osspkg.com/goppy/iosync v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/xtest v0.0.0-00010101000000-000000000000
)

require github.com/josharian/intern v1.0.0 // indirect
