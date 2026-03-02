SHELL=/bin/bash

run:
	go generate ./types/wsg.go
	go generate ./transport
	go run main.go --config=config.yaml

