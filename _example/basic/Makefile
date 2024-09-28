SHELL=/bin/bash

run:
	go run main.go --config=config.yaml

check:
	time ab -n 1230 -c 5 http://127.0.0.1:10000/users
	curl -v http://127.0.0.1:12000/metrics
