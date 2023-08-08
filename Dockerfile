FROM golang:1.20.7-alpine3.18@sha256:b255d9352ff4967f0424731acf5a262c3bb4b5bd712e6324ad9ed533c9d2aee7 AS build
WORKDIR /usr/src/app
RUN apk --no-cache add git make
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN make install

FROM golang:1.20.7-alpine3.18@sha256:b255d9352ff4967f0424731acf5a262c3bb4b5bd712e6324ad9ed533c9d2aee7
# When running as non-root user, GOCACHE must be set to a directory
# that is writable by that user. It will otherwise default to /.cache/go-build,
# which is owned by root.
# https://github.com/golang/go/issues/26280#issuecomment-445294378
ENV GOCACHE=/tmp/go-build
COPY --from=build /go/bin/cyclonedx-gomod /usr/local/bin/
USER 1000
ENTRYPOINT ["cyclonedx-gomod"]
CMD ["-h"]
