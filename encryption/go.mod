module go.osspkg.com/goppy/encryption

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/random => ../random
	go.osspkg.com/goppy/syscall => ../syscall
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.3.0
	go.osspkg.com/goppy/random v0.3.0
	go.osspkg.com/goppy/xtest v0.3.0
	golang.org/x/crypto v0.20.0
)

require go.osspkg.com/goppy/syscall v0.3.0 // indirect
