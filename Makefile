SHELL := /bin/bash

## help: show this help message
help:
	@ echo -e "Usage: make [target]\n"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## test: run unit tests
test:
	@ go test -race -v ./... -count=1