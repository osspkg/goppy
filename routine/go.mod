module go.osspkg.com/goppy/routine

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/iosync => ../iosync
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.1.2
	go.osspkg.com/goppy/iosync v0.1.5
)

require go.osspkg.com/goppy/syscall v0.1.3 // indirect
