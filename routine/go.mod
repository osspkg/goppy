module go.osspkg.com/goppy/routine

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.3.1
	go.osspkg.com/goppy/iosync v0.3.0
)

require go.osspkg.com/goppy/syscall v0.3.0 // indirect
