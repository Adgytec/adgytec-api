.PHONY: run build test prepareTest

run:
	go run cmd/server/main.go cmd/server/init.go

build:
	go build cmd/server/main.go cmd/server/init.go

test:
	go test -v ./...

prepareTest:
	go run ./test/prepare/main.go