.PHONY: run test build

export GOTOOLCHAIN := local

run:
	go run ./cmd/training-sync

test:
	go test -count=1 ./...

build:
	CGO_ENABLED=0 go build ./...
