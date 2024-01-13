module go.osspkg.com/goppy/graphics

go 1.20

replace (
	go.osspkg.com/goppy/errors => ../errors
	go.osspkg.com/goppy/syscall => ../syscall
)

require (
	go.osspkg.com/goppy/errors v0.3.0
	golang.org/x/image v0.15.0
)

require go.osspkg.com/goppy/syscall v0.3.0 // indirect
