module go.osspkg.com/goppy/ioutil

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/syscall => ../syscall
)

require go.osspkg.com/goppy/errors v0.1.2

require go.osspkg.com/goppy/syscall v0.1.3 // indirect
