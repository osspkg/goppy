SHELL=/bin/bash

help:
	go run main.go --help
	go run main.go env --help

run:
	go run main.go --config=config.yaml

show_env:
	go run main.go env

check:
	curl -v http://127.0.0.1:10000/users && echo "" && echo "" && echo ""
	time ab -n 1230 -c 5 http://127.0.0.1:10000/users && echo "" && echo "" && echo ""
	curl -v http://127.0.0.1:12000/metrics && echo "" && echo "" && echo ""
