SHELL:=/bin/bash

.PHONY: install
install:
	go install ./cmd/goppy

.PHONY: setup
setup:
	goppy setup-lib

.PHONY: lint
lint:
	goppy lint
	cd _example && goppy lint

.PHONY: license
license:
	goppy license

.PHONY: build
build:
	goppy build --arch=amd64 --cgo

.PHONY: tests
tests:
	goppy test

.PHONY: pre-commit
pre-commit: install setup lint license build tests

.PHONY: ci
ci: pre-commit

.PHONY: tidy
tidy:
	go mod tidy -v

example-tb: install
	cd ./_example/web-server-gen && \
		go generate -run goppy ./... && \
		go generate -run easyjson ./...