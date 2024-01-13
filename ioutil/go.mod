module go.osspkg.com/goppy/ioutil

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/syscall => ../syscall
)

require go.osspkg.com/goppy/errors v0.2.0

require go.osspkg.com/goppy/syscall v0.2.0 // indirect
