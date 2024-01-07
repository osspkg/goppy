SHELL=/bin/bash

run:
	go run main.go --config=config.yaml

build_ssl:
	openssl req -new -sha256 -key test.key -out test.csr
	openssl x509 -req -sha256 -in test.csr -signkey test.key -out test.crt -days 3650

check_ssl:
	curl -vvI --insecure https://127.0.0.1:8080

check:
	curl -vvI --insecure http://127.0.0.1:8080