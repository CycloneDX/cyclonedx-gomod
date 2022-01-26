ifeq ($(VERSION),)
	VERSION=v0.0.0-$(shell git show -s --date=format:'%Y%m%d%H%M%S' --format=%cd HEAD)-$(shell git rev-parse HEAD | head -c 12)
endif

LDFLAGS="-s -w -X github.com/CycloneDX/cyclonedx-gomod/internal/version.Version=${VERSION}"

build:
	mkdir -p ./bin
	CGO_ENABLED=0 go build -v -ldflags=${LDFLAGS} -o ./bin/cyclonedx-gomod ./cmd/cyclonedx-gomod
.PHONY: build

install:
	CGO_ENABLED=0 go install -v -ldflags=${LDFLAGS} ./cmd/cyclonedx-gomod
.PHONY: install

unit-test:
	go test -v -short -cover ./...
.PHONY: unit-test

test:
	go test -v -cover ./...
.PHONY: test

clean:
	go clean ./...
.PHONY: clean

docker:
	docker build -t cyclonedx/cyclonedx-gomod -f Dockerfile .
.PHONY: docker

goreleaser-dryrun:
	goreleaser release --skip-publish --snapshot
.PHONY: goreleaser-dryrun

examples-image: build
	docker build -t cyclonedx-gomod-examples -f Dockerfile.examples ./bin
.PHONY: examples-image

examples:
	docker run -i --rm -v "$(shell pwd)/examples:/examples" cyclonedx-gomod-examples
.PHONY: examples

all: clean build test
.PHONY: all
