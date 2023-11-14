module go.osspkg.com/goppy/iofile

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.1.0
	go.osspkg.com/goppy/xtest v0.1.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)
