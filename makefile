.PHONY: run build test

run:
	go run cmd/server/main.go cmd/server/init.go

build:
	go build cmd/server/main.go cmd/server/init.go

test:
	go test -v ./...