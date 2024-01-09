module go.osspkg.com/goppy/ioutil

go 1.18

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/syscall => ../syscall
)

require go.osspkg.com/goppy/errors v0.1.1

require go.osspkg.com/goppy/syscall v0.1.2 // indirect
