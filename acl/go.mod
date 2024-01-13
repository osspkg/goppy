module go.osspkg.com/goppy/acl

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.2.0
	go.osspkg.com/goppy/xtest v0.2.0
)

require go.osspkg.com/goppy/syscall v0.2.0 // indirect
