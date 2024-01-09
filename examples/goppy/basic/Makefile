SHELL=/bin/bash

run:
	go run main.go --config=config.yaml

show_metrics:
	curl -v http://127.0.0.1:12000/metrics

ab_test:
	time ab -n 1230 -c 5 http://127.0.0.1:8080/users