SHELL=/bin/bash

run:
	go run main.go --config=config.yaml

check:
	curl -vvv http://127.0.0.1:10000/oauth/r/ya && echo "" && echo "" && echo ""
	curl -vvv http://127.0.0.1:10000/oauth/r/google && echo "" && echo "" && echo ""