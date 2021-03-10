build:
	go build -v
.PHONY: build

test:
	go test -v -cover ./...
.PHONY: test

clean:
	go clean ./...
.PHONY: clean

bom:
	go mod tidy
	go mod download
	go run main.go

all: clean build test bench
.PHONY: all
