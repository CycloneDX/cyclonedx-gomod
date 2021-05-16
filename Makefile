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
	go run main.go -licenses -std -output bom.xml
	cyclonedx validate --input-file bom.xml --input-format xml --fail-on-errors
.PHONY: bom

goreleaser-dryrun:
	goreleaser release --skip-publish --snapshot
.PHONY: goreleaser-dryrun

all: clean build test
.PHONY: all
