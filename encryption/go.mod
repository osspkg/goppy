module go.osspkg.com/goppy/encryption

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/random => ../random
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.1.1
	go.osspkg.com/goppy/random v0.1.2
	go.osspkg.com/goppy/xtest v0.1.4
	golang.org/x/crypto v0.17.0
)
