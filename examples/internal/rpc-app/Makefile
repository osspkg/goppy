SHELL=/bin/bash

build_app:
	goppy build --main=./rpc-go-plugin/main.go,./rpc-unix-plugin/main.go,./main.go --cgo


run_app: build_app
	./build/app_amd64 --config=config.yaml --config-recovery

check:
	curl -iL http://127.0.0.1:8080/call/go-plugin && echo
	curl -iL http://127.0.0.1:8080/call/unix-plugin && echo