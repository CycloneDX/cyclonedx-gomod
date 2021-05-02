build:
	go build -v
.PHONY: build

test:
	go test -v -cover ./...
.PHONY: test

clean:
	go clean ./...
.PHONY: clean

docker:
	docker build -t cyclonedx/cyclonedx-gomod -f Dockerfile .
.PHONY: docker

bom:
	go mod download
	go mod tidy
	go run main.go -std
.PHONY: bom

all: clean build test bench
.PHONY: all
