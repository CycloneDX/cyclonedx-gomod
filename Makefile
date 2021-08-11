LDFLAGS="-s -w -X github.com/CycloneDX/cyclonedx-gomod/internal/version.Version=v0.0.0-$(shell git show -s --date=format:'%Y%m%d%H%M%S' --format=%cd HEAD)-$(shell git rev-parse HEAD | head -c 12)"

build:
	mkdir -p ./bin
	go build -v -ldflags=${LDFLAGS} -o ./bin/cyclonedx-gomod
.PHONY: build

install:
	go install -v -ldflags=${LDFLAGS}
.PHONY: install

generate:
	go generate -v ./...
.PHONY: generate

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

bom: build
	./bin/cyclonedx-gomod mod -licenses -std -json -output bom.json
	cyclonedx validate --input-file bom.json --fail-on-errors
.PHONY: bom

goreleaser-dryrun:
	goreleaser release --skip-publish --snapshot
.PHONY: goreleaser-dryrun

all: clean build test
.PHONY: all
