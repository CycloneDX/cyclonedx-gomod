GOFLAGS=-trimpath
LDFLAGS="-s -w"

build:
	mkdir -p ./bin
	CGO_ENABLED=0 go build -v ${GOFLAGS} -ldflags=${LDFLAGS} -o ./bin/cyclonedx-gomod ./cmd/cyclonedx-gomod
.PHONY: build

install:
	CGO_ENABLED=0 go install -v ${GOFLAGS} -ldflags=${LDFLAGS} ./cmd/cyclonedx-gomod
.PHONY: install

unit-test:
	go test -v -short -cover ./...
.PHONY: unit-test

test:
	go test -v -cover ./...
.PHONY: test

clean:
	rm -rf ./bin
	rm -rf ./dist
	go clean -testcache ./...
.PHONY: clean

docker:
	docker build -t cyclonedx/cyclonedx-gomod -f Dockerfile .
.PHONY: docker

goreleaser-dryrun:
	goreleaser release --skip-publish --skip-sign --snapshot
.PHONY: goreleaser-dryrun

build-examples-image:
	docker build -t cyclonedx/cyclonedx-gomod:examples -f Dockerfile.examples .
.PHONY: build-examples-image

examples: build-examples-image
	docker run -it --rm -v "$(shell pwd)/examples:/examples" cyclonedx/cyclonedx-gomod:examples
.PHONY: examples

all: clean build test
.PHONY: all
