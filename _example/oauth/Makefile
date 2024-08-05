SHELL=/bin/bash

run:
	go run main.go --config=config.yaml

check:
	curl -v http://127.0.0.1:10000/oauth/r/ya && echo ""
	curl -v http://127.0.0.1:10000/oauth/r/google && echo ""