module go.osspkg.com/goppy/acl

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.1.2
	go.osspkg.com/goppy/xtest v0.1.4
)

require go.osspkg.com/goppy/syscall v0.1.3 // indirect
