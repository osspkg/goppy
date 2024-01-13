module go.osspkg.com/goppy/metrics

go 1.20

replace (
	go.osspkg.com/goppy/env => ../env
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/ioutil => ../ioutil
	go.osspkg.com/goppy/plugins => ../plugins
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/web => ../web
	go.osspkg.com/goppy/xc => ../xc
	go.osspkg.com/goppy/xlog => ../xlog
	go.osspkg.com/goppy/xnet => ../xnet
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	github.com/prometheus/client_golang v1.18.0
	github.com/prometheus/client_model v0.5.0
	go.osspkg.com/goppy/env v0.3.0
	go.osspkg.com/goppy/plugins v0.3.0
	go.osspkg.com/goppy/syscall v0.3.0
	go.osspkg.com/goppy/web v0.3.0
	go.osspkg.com/goppy/xc v0.3.0
	go.osspkg.com/goppy/xlog v0.3.0
	go.osspkg.com/goppy/xnet v0.3.0
	go.osspkg.com/goppy/xtest v0.3.0
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	go.osspkg.com/goppy/errors v0.3.0 // indirect
	go.osspkg.com/goppy/iosync v0.3.0 // indirect
	go.osspkg.com/goppy/ioutil v0.3.0 // indirect
	go.osspkg.com/static v1.4.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
)
