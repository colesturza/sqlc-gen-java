.PHONY: build test

build:
	go build ./...

test:
	go test ./...

all: bin/sqlc-gen-java bin/sqlc-gen-java.wasm

bin/sqlc-gen-java: bin go.mod go.sum $(wildcard **/*.go)
	cd plugin && go build -o ../bin/sqlc-gen-java main.go

bin/sqlc-gen-java.wasm: bin/sqlc-gen-java
	cd plugin && GOOS=wasip1 GOARCH=wasm go build -o ../bin/sqlc-gen-java.wasm main.go

bin:
	mkdir -p bin