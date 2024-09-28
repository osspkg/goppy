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

.PHONY: license
license:
	goppy license

.PHONY: build
build:
	goppy build --arch=amd64

.PHONY: tests
tests:
	goppy test

.PHONY: ci
ci: install setup license lint build tests

.PHONY: tidy
tidy:
	go mod tidy -v
