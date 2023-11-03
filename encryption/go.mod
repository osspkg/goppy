module go.osspkg.com/goppy/encryption

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/random => ../random
	go.osspkg.com/goppy/xtest => ../xtest
)

require (
	go.osspkg.com/goppy/errors v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/random v0.0.0-00010101000000-000000000000
	go.osspkg.com/goppy/xtest v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.14.0
)
